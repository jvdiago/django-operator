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

// DjangoStaticSpec defines the desired state of DjangoStatic.
type DjangoStaticSpec struct{}

// DjangoStaticStatus defines the observed state of DjangoStatic.
type DjangoStaticStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Collected metav1.Time `json:"collected"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DjangoStatic is the Schema for the djangostatics API.
type DjangoStatic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DjangoStaticSpec   `json:"spec,omitempty"`
	Status DjangoStaticStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DjangoStaticList contains a list of DjangoStatic.
type DjangoStaticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DjangoStatic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DjangoStatic{}, &DjangoStaticList{})
}
