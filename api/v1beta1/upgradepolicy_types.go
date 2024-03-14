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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UpgradePolicySpec defines the desired state of UpgradePolicy
type UpgradePolicySpec struct {
	NodeUpgradeTemplateReference *corev1.ObjectReference `json:"nodeUpgradeTemplateRef"`
}

// UpgradePolicyStatus defines the observed state of UpgradePolicy
type UpgradePolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UpgradePolicy is the Schema for the upgradepolicies API
type UpgradePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpgradePolicySpec   `json:"spec,omitempty"`
	Status UpgradePolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpgradePolicyList contains a list of UpgradePolicy
type UpgradePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UpgradePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UpgradePolicy{}, &UpgradePolicyList{})
}
