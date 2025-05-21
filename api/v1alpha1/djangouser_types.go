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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DjangoUserSpec defines the desired state of DjangoUser.
type DjangoUserSpec struct {
	Username          string            `json:"username"`
	Email             string            `json:"email,omitempty"`
	PasswordSecretRef SecretKeySelector `json:"passwordSecretRef"`
	Superuser         bool              `json:"superuser"`
}

type SecretKeySelector struct {
	// Name of the Secret in the same namespace
	Name string `json:"name"`
	// Key within Data
	Key string `json:"key"`
}

// DjangoUserStatus defines the observed state of DjangoUser.
type DjangoUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Created metav1.Time `json:"created"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DjangoUser is the Schema for the djangousers API.
type DjangoUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DjangoUserSpec   `json:"spec,omitempty"`
	Status DjangoUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DjangoUserList contains a list of DjangoUser.
type DjangoUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DjangoUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DjangoUser{}, &DjangoUserList{})
}
