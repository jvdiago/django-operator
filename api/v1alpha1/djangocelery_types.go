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

// DjangoCelerySpec defines the desired state of DjangoCelery.
type DjangoCelerySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of DjangoCelery. Edit djangocelery_types.go to remove/update
	App    string `json:"app"`
	Worker string `json:"worker,omitempty"`
	Task   string `json:"task,omitempty"`
}

// DjangoCeleryStatus defines the observed state of DjangoCelery.
type DjangoCeleryStatus struct {
	Executed metav1.Time `json:"executed,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DjangoCelery is the Schema for the djangoceleries API.
type DjangoCelery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DjangoCelerySpec   `json:"spec,omitempty"`
	Status DjangoCeleryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DjangoCeleryList contains a list of DjangoCelery.
type DjangoCeleryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DjangoCelery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DjangoCelery{}, &DjangoCeleryList{})
}
