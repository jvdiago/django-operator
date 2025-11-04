package controller

import (
	"context"
	"encoding/json"
	"fmt"
	charts "github.com/jvdiago/django-helm-template"
	"github.com/operator-framework/helm-operator-plugins/pkg/reconciler"
	"github.com/operator-framework/helm-operator-plugins/pkg/values"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
)

// specTranslator reads spec.Values into chartutil.Values
func specTranslator(c client.Client) values.Translator {
	return values.TranslatorFunc(func(ctx context.Context, u *unstructured.Unstructured) (chartutil.Values, error) {
		// convert Unstructured → typed CR
		app := &djangov1alpha1.DjangoApp{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, app); err != nil {
			return nil, err
		}
		if app.Spec.Values == nil {
			return chartutil.Values{}, nil
		}
		// return exactly what the user put under `.spec.values`
		var m map[string]interface{}
		if err := json.Unmarshal(app.Spec.Values.Raw, &m); err != nil {
			return nil, err
		}

		// 4) Wrap and return
		return chartutil.Values(m), nil
	})
}

// +kubebuilder:rbac:groups=apps.django.djangooperator,resources=djangoapps,verbs=get;list;watch;create;update;patch;delete
// SetupHelmController wires the generic Helm-based reconciler into the manager.
func SetupHelmController(mgr ctrl.Manager) error {
	// Load the embedded chart
	chartObj, err := charts.Chart()
	if err != nil {
		return fmt.Errorf("failed to load embedded chart: %w", err)
	}
	// chartObj, err := loader.Load("helm-charts/djangoapp.tar.gz")
	// if err != nil {
	// 	panic(err)
	// }
	r, err := reconciler.New(
		reconciler.WithChart(*chartObj),
		reconciler.WithGroupVersionKind(schema.GroupVersionKind{
			Group:   "django.djangooperator",
			Version: "v1alpha1",
			Kind:    "DjangoApp",
		}),
		reconciler.SkipDependentWatches(true),
		reconciler.WithMaxConcurrentReconciles(1),
		reconciler.SkipPrimaryGVKSchemeRegistration(true),
		reconciler.WithValueTranslator(specTranslator(mgr.GetClient())),
		reconciler.WithLog(logf.Log.WithName("helm").WithName("DjangoApp")),
	)
	if err != nil {
		return fmt.Errorf("creating helm reconciler: %w", err)
	}

	if err := r.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("setting up helm reconciler: %w", err)
	}

	return nil
}
