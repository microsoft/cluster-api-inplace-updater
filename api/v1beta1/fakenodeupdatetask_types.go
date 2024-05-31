package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// +kubebuilder:object:root=true
type FakeNodeUpdateTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FakeNodeUpdateTaskSpec   `json:"spec,omitempty"`
	Status FakeNodeUpdateTaskStatus `json:"status,omitempty"`
}

type FakeNodeUpdateTaskSpec struct {
	Field1         string          `json:"field1,omitempty"`
	Field2         string          `json:"field2,omitempty"`
	Field3         string          `json:"field3,omitempty"`
	MachineName    string          `json:"machineName,omitempty"`
	NewMachineSpec MachineSpec     `json:"newMachineSpec,omitempty"`
	Phase          UpdateTaskPhase `json:"phase,omitempty"`
}

type FakeNodeUpdateTaskStatus struct {
	State      UpdateTaskState       `json:"state,omitempty"`
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
type FakeNodeUpdateTaskTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FakeNodeUpdateTaskTemplateSpec `json:"spec,omitempty"`
}

type FakeNodeUpdateTaskTemplateSpec struct {
	Template FakeNodeUpdateTaskTemplateResource `json:"template"`
}

type FakeNodeUpdateTaskTemplateResource struct {
	ObjectMeta clusterv1.ObjectMeta   `json:"metadata,omitempty"`
	Spec       FakeNodeUpdateTaskSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// FakeNodeUpdateTaskList contains a list of FakeNodeUpdateTask
type FakeNodeUpdateTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FakeNodeUpdateTask `json:"items"`
}

// +kubebuilder:object:root=true
// FakeNodeUpdateTaskTemplateList contains a list of FakeNodeUpdateTaskTemplate
type FakeNodeUpdateTaskTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FakeNodeUpdateTaskTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FakeNodeUpdateTask{}, &FakeNodeUpdateTaskTemplate{}, &FakeNodeUpdateTaskList{}, &FakeNodeUpdateTaskTemplateList{})
}
