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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// DockerNodeUpdateTaskSpec defines the desired state of DockerNodeUpdateTask
type DockerNodeUpdateTaskSpec struct {
	MachineName    string          `json:"machineName,omitempty"`
	NewMachineSpec MachineSpec     `json:"newMachineSpec,omitempty"`
	Phase          UpdateTaskPhase `json:"phase,omitempty"`
}

// DockerNodeUpdateTaskStatus defines the observed state of DockerNodeUpdateTask
type DockerNodeUpdateTaskStatus struct {
	State      UpdateTaskState       `json:"state,omitempty"`
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DockerNodeUpdateTask is the Schema for the dockernodeupdatetasks API
type DockerNodeUpdateTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerNodeUpdateTaskSpec   `json:"spec,omitempty"`
	Status DockerNodeUpdateTaskStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DockerNodeUpdateTaskList contains a list of DockerNodeUpdateTask
type DockerNodeUpdateTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerNodeUpdateTask `json:"items"`
}

type DockerNodeUpdateTaskTemplateSpec struct {
	Template DockerNodeUpdateTaskTemplateResource `json:"template"`
}

type DockerNodeUpdateTaskTemplateResource struct {
	ObjectMeta clusterv1.ObjectMeta     `json:"metadata,omitempty"`
	Spec       DockerNodeUpdateTaskSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type DockerNodeUpdateTaskTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DockerNodeUpdateTaskTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// DockerNodeUpdateTaskTemplateList contains a list of DockerNodeUpdateTaskTemplate
type DockerNodeUpdateTaskTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerNodeUpdateTaskTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerNodeUpdateTask{}, &DockerNodeUpdateTaskList{}, &DockerNodeUpdateTaskTemplate{}, &FakeNodeUpdateTaskTemplateList{})
}
