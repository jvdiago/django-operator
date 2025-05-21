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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// DjangoUserReconciler reconciles a DjangoUser object
type DjangoUserReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	RESTCfg   *rest.Config
	Clientset *kubernetes.Clientset
}

// As our operator us confined in a namespace, the role file needs to be edited manually. Nevertheless,
// all the annotation are left for when the gen binary supports Role and not only ClusterRole

// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:rbac:groups=django.my.domain,resources=djangousers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=django.my.domain,resources=djangousers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=django.my.domain,resources=djangousers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DjangoUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	// 1) Fetch the DjangoUser
	var du djangov1alpha1.DjangoUser
	if err := r.Get(ctx, req.NamespacedName, &du); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted, nothing to do
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// 2) Skip if already created
	if !du.Status.Created.IsZero() {
		return ctrl.Result{}, nil
	}
	// 3) Find the Django pod in this namespace
	podList := &corev1.PodList{}
	sel := labels.SelectorFromSet(labels.Set{
		"app.kubernetes.io/component": "django-server",
	})
	if err := r.List(ctx, podList, &client.ListOptions{
		Namespace:     req.Namespace,
		LabelSelector: sel,
	}); err != nil {
		return ctrl.Result{}, err
	}
	if len(podList.Items) == 0 {
		logger.Info("no django-server pod found; retrying shortly")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	pod := podList.Items[0]

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
u.save()`, du.Spec.Username, du.Spec.Password, du.Spec.Email, pySuperuser),
	}

	// 5) Exec command
	if err := r.execInPod(ctx, &pod, shellCmd); err != nil {
		logger.Error(err, "failed to exec create user command", du.Spec.Username, "pod", pod.Name)
		return ctrl.Result{}, err
	}

	// 6) Update status.Created
	du.Status.Created = metav1.Now()
	if err := r.Status().Update(ctx, &du); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("User created", "user", du.Spec.Username, "superuser", du.Spec.Superuser)

	return ctrl.Result{}, nil
}

// execInPod runs the given command in the first container of the pod
func (r *DjangoUserReconciler) execInPod(ctx context.Context, pod *corev1.Pod, command []string) error {
	req := r.Clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   command,
			Container: pod.Spec.Containers[0].Name,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(r.RESTCfg, "POST", req.URL())
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *DjangoUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// initialize REST config & clientset
	r.RESTCfg = mgr.GetConfig()
	cs, err := kubernetes.NewForConfig(r.RESTCfg)
	if err != nil {
		return err
	}
	r.Clientset = cs

	return ctrl.NewControllerManagedBy(mgr).
		For(&djangov1alpha1.DjangoUser{}).
		Named("djangouser").
		Complete(r)
}
