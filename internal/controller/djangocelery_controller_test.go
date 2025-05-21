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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
)

var _ = Describe("DjangoCelery Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		djangocelery := &djangov1alpha1.DjangoCelery{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind DjangoCelery")
			err := k8sClient.Get(ctx, typeNamespacedName, djangocelery)
			if err != nil && errors.IsNotFound(err) {
				resource := &djangov1alpha1.DjangoCelery{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: djangov1alpha1.DjangoCelerySpec{
						App:    "app",
						Worker: "worker@celery",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &djangov1alpha1.DjangoCelery{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DjangoCelery")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DjangoCeleryReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Pods:   testPodRunner{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Verifying the status.executed timestamp is set")
			// Re-fetch the resource
			updated := &djangov1alpha1.DjangoCelery{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())

			// Ensure Status.Created is non-zero
			Expect(updated.Status.Executed.IsZero()).To(BeFalse(), "expected Status.Executed to be set")

		})
	})
})
