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

	"github.com/microsoft/cluster-api-inplace-upgrader/stub"
)

// UpgradeTaskSpec defines the desired state of UpgradeTask
type UpgradeTaskSpec struct {
	ClusterRef             *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	ControlPlaneRef        *corev1.ObjectReference  `json:"controlPlaneRef,omitempty"`
	MachineDeploymentRef   *corev1.ObjectReference  `json:"machineDeploymentRef,omitempty"`
	MachinesRequireUpgrade []corev1.ObjectReference `json:"machinesRequireUpgrade,omitempty"`
	NewMachineSpec         stub.MachineSpec         `json:"newMachineSpec,omitempty"`

	TargetPhase UpgradeTaskPhase `json:"phase,omitempty"`
}

// UpgradeTaskStatus defines the observed state of UpgradeTask
type UpgradeTaskStatus struct {
	Phase UpgradeTaskPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UpgradeTask is the Schema for the upgradetasks API
type UpgradeTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpgradeTaskSpec   `json:"spec,omitempty"`
	Status UpgradeTaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpgradeTaskList contains a list of UpgradeTask
type UpgradeTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UpgradeTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UpgradeTask{}, &UpgradeTaskList{})
}
