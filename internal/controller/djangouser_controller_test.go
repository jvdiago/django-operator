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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	djangov1alpha1 "github.com/jvdiago/django-operator/api/v1alpha1"
)

// ------------------------------------------------------------------
// TEST‐DOUBLE DECLARATIONS (top‐level, not inside Describe)
// ------------------------------------------------------------------

var _ = Describe("DjangoUser Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		djangouser := &djangov1alpha1.DjangoUser{}

		BeforeEach(func() {
			pwSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pw-" + resourceName,
					Namespace: typeNamespacedName.Namespace,
				},
				StringData: map[string]string{
					"password": "S3cr3t",
				},
			}
			Expect(k8sClient.Create(ctx, pwSecret)).To(Succeed())

			By("creating the custom resource for the Kind DjangoUser")
			err := k8sClient.Get(ctx, typeNamespacedName, djangouser)
			if err != nil && errors.IsNotFound(err) {
				resource := &djangov1alpha1.DjangoUser{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: djangov1alpha1.DjangoUserSpec{
						Username: resourceName,
						Email:    "test@example.com",
						PasswordSecretRef: djangov1alpha1.SecretKeySelector{
							Name: pwSecret.Name,
							Key:  "password",
						},
						Superuser: false,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &djangov1alpha1.DjangoUser{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DjangoUser")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			pwSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pw-" + resourceName,
					Namespace: typeNamespacedName.Namespace,
				},
			}
			Expect(k8sClient.Delete(ctx, pwSecret)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			tr := &DjangoUserReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err := tr.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
