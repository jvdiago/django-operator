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
	"time"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DjangoMigrateReconciler reconciles a DjangoMigrate object
type DjangoMigrateReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Pods           PodRunner
	DjangoPodlabel PodLabel
	KeepCRs        int
}

// +kubebuilder:rbac:groups=django.djangooperator,resources=djangomigrates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangomigrates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangomigrates/finalizers,verbs=update

// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DjangoMigrateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	// Fetch the DjangoUser
	var dm djangov1alpha1.DjangoMigrate
	if err := r.Get(ctx, req.NamespacedName, &dm); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted, nothing to do
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Skip if already created
	if !dm.Status.Applied.IsZero() {
		return ctrl.Result{}, nil
	}
	pod, err := r.Pods.FindDjangoPod(ctx, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if pod == nil {
		logger.Info("no django-server pod found; retrying shortly")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	// Build the command
	shellCmd := []string{
		"python", "manage.py", "migrate", "--noinput",
	}
	if dm.Spec.Fake {
		shellCmd = append(shellCmd, "--fake")
	}
	if dm.Spec.App != "" {
		shellCmd = append(shellCmd, dm.Spec.App)
	}
	if dm.Spec.Migration != "" {
		shellCmd = append(shellCmd, dm.Spec.Migration)
	}
	// Exec command
	if err := r.Pods.ExecInPod(ctx, pod, shellCmd); err != nil {
		logger.Error(err, "failed to exec migrate command", shellCmd, "pod", pod.Name)
		return ctrl.Result{}, err
	}

	// Update status.Applied
	dm.Status.Applied = metav1.Now()
	if err := r.Status().Update(ctx, &dm); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Migration applied", "migrate", dm.Name)

	// keep only the most-recent DjangoMigrate objects
	migrateGVK := djangov1alpha1.GroupVersion.WithKind("DjangoMigrate")
	if err := pruneOldCRs(r.Client, ctx, migrateGVK, req.Namespace, r.KeepCRs); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *DjangoMigrateReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&djangov1alpha1.DjangoMigrate{}).
		Named("djangomigrate").
		Complete(r)
}
