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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const UpdateTaskFinalizer = "updatetask.update.extension.cluster.x-k8s.io"

const (
	UpdateOperationCondition clusterv1.ConditionType = "UpdateOperation"
	AbortOperationCondition  clusterv1.ConditionType = "AbortOperation"
	TrackOperationCondition  clusterv1.ConditionType = "TrackOperation"
)

const (
	NodeUpdateFailedReason       = "UpdateFailed"
	PreflightCheckFailedReason   = "PreflightCheckFailed"
	CreateNodeUpdateFailedReason = "CreateNodeUpdateFailed"

	AbortNodeUpdateFailedReason = "AbortNodeUpdateFailed"
)

// UpdateTaskSpec defines the desired state of UpdateTask
type UpdateTaskSpec struct {
	// ClusterRef reference to capi Cluster
	ClusterRef *corev1.ObjectReference `json:"clusterRef,omitempty"`

	// ControlPlaneRef reference to capi controlplane cr when this is an update on controlplane
	// +optional
	ControlPlaneRef *corev1.ObjectReference `json:"controlPlaneRef,omitempty"`

	// MachineDeploymentRef reference to capi machinedeployment cr when this is an update on machinedeployment
	// +optional
	MachineDeploymentRef *corev1.ObjectReference `json:"machineDeploymentRef,omitempty"`

	// MachinesRequireUpdate reference to list of machines which need to be updated
	// +optional
	MachinesRequireUpdate []corev1.ObjectReference `json:"machinesRequireUpdate,omitempty"`

	// NewMachineSpec indicates the new machine spec to be updated to
	NewMachineSpec MachineSpec `json:"newMachineSpec,omitempty"`

	// NodeUpdateTemplate indicates the template to be used for creating nodeUpdateTask
	NodeUpdateTemplate NodeUpdateTaskTemplateSpec `json:"template,omitempty"`

	// TargetPhase indicates the phase of update
	TargetPhase UpdateTaskPhase `json:"phase,omitempty"`
}

// NodeUpdateTemplateSpec is template of nodeUpdateTask
type NodeUpdateTaskTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	InfrastructureRef corev1.ObjectReference `json:"infrastructureRef"`
}

// UpdateTaskStatus defines the observed state of UpdateTask
type UpdateTaskStatus struct {
	// State indicates update state
	State UpdateTaskState `json:"state,omitempty"`

	// Conditions defines current service state of the updateTask.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
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

// GetConditions returns the set of conditions for this object.
func (c *UpdateTask) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *UpdateTask) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
}
