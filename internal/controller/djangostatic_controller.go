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
	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// DjangoStaticReconciler reconciles a DjangoStatic object
type DjangoStaticReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Pods           PodRunner
	DjangoPodlabel PodLabel
}

// +kubebuilder:rbac:groups=django.djangooperator,resources=djangostatics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangostatics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=django.djangooperator,resources=djangostatics/finalizers,verbs=update

// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *DjangoStaticReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	// Fetch the DjangoUser
	var ds djangov1alpha1.DjangoStatic
	if err := r.Get(ctx, req.NamespacedName, &ds); err != nil {
		if errors.IsNotFound(err) {
			// CR deleted, nothing to do
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Skip if already created
	if !ds.Status.Collected.IsZero() {
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
	// 4) Build the command
	shellCmd := []string{
		"python", "manage.py", "collectstatic", "--noinput",
	}
	// Exec command
	if err := r.Pods.ExecInPod(ctx, pod, shellCmd); err != nil {
		logger.Error(err, "failed to exec collectstatic command", shellCmd, "pod", pod.Name)
		return ctrl.Result{}, err
	}

	// Update status.Applied
	ds.Status.Collected = metav1.Now()
	if err := r.Status().Update(ctx, &ds); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Statics collected", "collectstatic", ds.Name)

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *DjangoStaticReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&djangov1alpha1.DjangoStatic{}).
		Named("djangostatic").
		Complete(r)
}
