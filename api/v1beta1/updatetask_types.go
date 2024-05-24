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

// UpdateTaskSpec defines the desired state of UpdateTask
type UpdateTaskSpec struct {
	ClusterRef            *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	ControlPlaneRef       *corev1.ObjectReference  `json:"controlPlaneRef,omitempty"`
	MachineDeploymentRef  *corev1.ObjectReference  `json:"machineDeploymentRef,omitempty"`
	MachinesRequireUpdate []corev1.ObjectReference `json:"machinesRequireUpdate,omitempty"`
	NewMachineSpec        MachineSpec              `json:"newMachineSpec,omitempty"`

	TargetPhase UpdateTaskPhase `json:"phase,omitempty"`
}

// UpdateTaskStatus defines the observed state of UpdateTask
type UpdateTaskStatus struct {
	Phase UpdateTaskPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UpdateTask is the Schema for the updatetasks API
type UpdateTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UpdateTaskSpec   `json:"spec,omitempty"`
	Status UpdateTaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UpdateTaskList contains a list of UpdateTask
type UpdateTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UpdateTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UpdateTask{}, &UpdateTaskList{})
}
