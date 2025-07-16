/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DjangoUserReconciler reconciles a DjangoUser object

type DjangoUserReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Pods           PodRunner
	DjangoPodlabel PodLabel
	KeepCRs        int
}

// As our operator us confined in a namespace, the role file needs to be edited manually. Nevertheless,
// all the annotation are left for when the gen binary supports Role and not only ClusterRole

// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangousers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangousers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangousers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DjangoUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	// Fetch the DjangoUser
	var du djangov1alpha1.DjangoUser
	if err := r.Get(ctx, req.NamespacedName, &du); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted, nothing to do
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Skip if already created
	if !du.Status.Created.IsZero() {
		return ctrl.Result{}, nil
	}
	// Read the password from the Secret
	var pwSecret corev1.Secret
	if err := r.Get(ctx,
		types.NamespacedName{Namespace: req.Namespace, Name: du.Spec.PasswordSecretRef.Name},
		&pwSecret,
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("reading password secret: %w", err)
	}
	raw, ok := pwSecret.Data[du.Spec.PasswordSecretRef.Key]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("secret %s missing key %q",
			du.Spec.PasswordSecretRef.Name, du.Spec.PasswordSecretRef.Key)
	}
	password := string(raw)
	pod, err := r.Pods.FindDjangoPod(ctx, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if pod == nil {
		logger.Info("no django-server pod found; retrying shortly")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	// 4) Build the command
	pySuperuser := "False"
	if du.Spec.Superuser {
		pySuperuser = "True"
	}
	shellCmd := []string{
		"python", "manage.py", "shell", "-c",
		fmt.Sprintf(`
from django.contrib.auth import get_user_model;
User = get_user_model();
username = '%s';
password = '%s';
email = '%s';
superuser = '%s';
u, created = User.objects.get_or_create(username=username, defaults={
    "email": email,
    "is_staff": True,
    "is_superuser": superuser,
    "is_active": True
});
if not created:
    u.email = email
    u.is_staff = True
    u.is_superuser = superuser
    u.is_active = True
u.set_password(password);
u.save()`, du.Spec.Username, password, du.Spec.Email, pySuperuser),
	}

	// Exec command
	if err := r.Pods.ExecInPod(ctx, pod, shellCmd); err != nil {
		logger.Error(err, "failed to exec create user command", du.Spec.Username, "pod", pod.Name)
		return ctrl.Result{}, err
	}

	// Update status.Created
	du.Status.Created = metav1.Now()
	if err := r.Status().Update(ctx, &du); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("User created", "user", du.Spec.Username, "superuser", du.Spec.Superuser)

	// keep only the most-recent DjangoUser objects
	userGVK := djangov1alpha1.GroupVersion.WithKind("DjangoCelery")
	if err := pruneOldCRs(r.Client, ctx, userGVK, req.Namespace, r.KeepCRs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DjangoUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// initialize REST config & clientset
	RestCFG := mgr.GetConfig()
	cs, err := kubernetes.NewForConfig(RestCFG)
	if err != nil {
		return err
	}

	// wire in the real PodRunner
	r.Pods = DjangoPodRunner{
		Client:    r.Client,
		RESTCfg:   RestCFG,
		Clientset: cs,
		Label:     r.DjangoPodlabel,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&djangov1alpha1.DjangoUser{}).
		Named("djangouser").
		Complete(r)
}
