package controller

import (
	"context"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// listCRs returns all objects of the given GVK in namespace, sorted newest-first.
func listCRs(
	c client.Client,
	ctx context.Context,
	gvk schema.GroupVersionKind,
	namespace string,
) ([]unstructured.Unstructured, error) {
	// build an UnstructuredList of the right kind
	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(gvk)

	if err := c.List(ctx, ul, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	// sort descending by CreationTimestamp
	sort.Slice(ul.Items, func(i, j int) bool {
		return ul.Items[i].GetCreationTimestamp().Time.After(
			ul.Items[j].GetCreationTimestamp().Time,
		)
	})

	return ul.Items, nil
}

// pruneOldCRs deletes everything except the `keep` newest instances of the given GVK.
func pruneOldCRs(
	c client.Client,
	ctx context.Context,
	gvk schema.GroupVersionKind,
	namespace string,
	keep int,
) error {

	logger := logf.FromContext(ctx)
	items, err := listCRs(c, ctx, gvk, namespace)
	if err != nil {
		return err
	}

	// 0 means do not delete CRs
	if keep < 1 {
		return nil
	}

	if len(items) <= keep {
		return nil
	}

	// delete the ones after the first `keep`
	for _, u := range items[keep:] {
		if err := c.Delete(ctx, &u); err != nil {
			return err
		}
		logger.Info(
			"Deleted old CR",
			"kind", gvk.Kind,
			"name", u.GetName())
	}
	return nil
}
