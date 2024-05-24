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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
)

type MachineSpec struct {
	Machine         *clusterv1.Machine         `json:"machine,omitempty"`
	BootstrapConfig *unstructured.Unstructured `json:"bootstrapConfig,omitempty"`
	InfraMachine    *unstructured.Unstructured `json:"infraMachine,omitempty"`
}

func ControlPlaneExternalUpdate(*ControlPlaneExternalUpdateRequest, *ControlPlaneExternalUpdateResponse) {
}

func MachineDeploymentExternalUpdate(*MachineDeploymentExternalUpdateRequest, *MachineDeploymentExternalUpdateResponse) {
}

// ControlPlaneExternalUpdateRequest is the input to an external update
// strategy implementer for a CP.
type ControlPlaneExternalUpdateRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	runtimehooksv1.CommonRequest `json:",inline"`
	ClusterRef                   *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	ControlPlaneRef              *corev1.ObjectReference  `json:"controlPlaneRef,omitempty"`
	MachinesRequireUpdate        []corev1.ObjectReference `json:"machinesRequireUpdate,omitempty"`
	NewMachine                   MachineSpec              `json:"newMachine,omitempty"`
}

// ControlPlaneExternalUpdateResponse is the response from an external
// update strategy implementer.
type ControlPlaneExternalUpdateResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	runtimehooksv1.CommonRetryResponse `json:",inline"`
}

// MachineDeploymentExternalUpdateRequest is the input to an external update
// strategy implementer for a groups of machines.
type MachineDeploymentExternalUpdateRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	runtimehooksv1.CommonRequest `json:",inline"`
	ClusterRef                   *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	MachineDeploymentRef         *corev1.ObjectReference  `json:"machineDeploymentRef,omitempty"`
	MachinesRequireUpdate        []corev1.ObjectReference `json:"machinesRequireUpdate,omitempty"`
	NewMachine                   MachineSpec              `json:"newMachine,omitempty"`
}

// MachineDeploymentExternalUpdateResponse is the response from an external
// update strategy implementer.
type MachineDeploymentExternalUpdateResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	runtimehooksv1.CommonRetryResponse `json:",inline"`
}
