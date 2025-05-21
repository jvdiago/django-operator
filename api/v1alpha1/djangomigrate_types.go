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

// DjangoMigrateSpec defines the desired state of DjangoMigrate.
type DjangoMigrateSpec struct {
	Fake      bool   `json:"fake,omitempty"`
	App       string `json:"app,omitempty"`
	Migration string `json:"migration,omitempty"`
}

// DjangoMigrateStatus defines the observed state of DjangoMigrate.
type DjangoMigrateStatus struct {
	Applied metav1.Time `json:"applied"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DjangoMigrate is the Schema for the djangomigrates API.
type DjangoMigrate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DjangoMigrateSpec   `json:"spec,omitempty"`
	Status DjangoMigrateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DjangoMigrateList contains a list of DjangoMigrate.
type DjangoMigrateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DjangoMigrate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DjangoMigrate{}, &DjangoMigrateList{})
}
