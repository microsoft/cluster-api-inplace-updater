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

package stub

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

func ControlPlaneExternalUpgrade(*ControlPlaneExternalUpgradeRequest, *ControlPlaneExternalUpgradeResponse) {
}

func MachineDeploymentExternalUpgrade(*MachineDeploymentExternalUpgradeRequest, *MachineDeploymentExternalUpgradeResponse) {
}

// ControlPlaneExternalUpgradeRequest is the input to an external upgrade
// strategy implementer for a CP.
type ControlPlaneExternalUpgradeRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	runtimehooksv1.CommonRequest `json:",inline"`
	ClusterRef                   *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	ControlPlaneRef              *corev1.ObjectReference  `json:"controlPlaneRef,omitempty"`
	MachinesRequireUpgrade       []corev1.ObjectReference `json:"machinesRequireUpgrade,omitempty"`
	NewMachine                   MachineSpec              `json:"newMachine,omitempty"`
}

// ControlPlaneExternalUpgradeResponse is the response from an external
// upgrade strategy implementer.
type ControlPlaneExternalUpgradeResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	runtimehooksv1.CommonRetryResponse `json:",inline"`
}

// MachineDeploymentExternalUpgradeRequest is the input to an external upgrade
// strategy implementer for a groups of machines.
type MachineDeploymentExternalUpgradeRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	runtimehooksv1.CommonRequest `json:",inline"`
	ClusterRef                   *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	MachineDeploymentRef         *corev1.ObjectReference  `json:"machineDeploymentRef,omitempty"`
	MachinesRequireUpgrade       []corev1.ObjectReference `json:"machinesRequireUpgrade,omitempty"`
	NewMachine                   MachineSpec              `json:"newMachine,omitempty"`
}

// MachineDeploymentExternalUpgradeResponse is the response from an external
// upgrade strategy implementer.
type MachineDeploymentExternalUpgradeResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	runtimehooksv1.CommonRetryResponse `json:",inline"`
}