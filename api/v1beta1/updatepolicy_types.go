/*
Copyright 2024.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdatePolicySpec defines the desired state of UpdatePolicy
type UpdatePolicySpec struct {
	NodeUpdateTemplateRef *corev1.ObjectReference `json:"nodeUpdateTemplateRef"`
}

// UpdatePolicyStatus defines the observed state of UpdatePolicy
type UpdatePolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UpdatePolicy is the Schema for the updatepolicies API
type UpdatePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpdatePolicySpec   `json:"spec,omitempty"`
	Status UpdatePolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpdatePolicyList contains a list of UpdatePolicy
type UpdatePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UpdatePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UpdatePolicy{}, &UpdatePolicyList{})
}
