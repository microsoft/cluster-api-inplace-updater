package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MachineSpec struct {
	Machine         *clusterv1.Machine
	BootstrapConfig *unstructured.Unstructured
	InfraMachine    *unstructured.Unstructured
}

func ControlPlaneExternalUpgrade(*ControlPlaneExternalUpgradeRequest, *ControlPlaneExternalUpgradeResponse) {
}

func MachinesExternalUpgrade(*MachinesExternalUpgradeRequest, *MachinesExternalUpgradeResponse) {
}

// ControlPlaneExternalUpgradeRequest is the input to an external upgrade
// strategy implementer for a CP.
type ControlPlaneExternalUpgradeRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	CommonRequest          `json:",inline"`
	Cluster                corev1.ObjectReference
	ControlPlane           corev1.ObjectReference
	MachinesRequireUpgrade []corev1.ObjectReference
	NewMachine             *MachineSpec
}

// ControlPlaneExternalUpgradeResponse is the response from an external
// upgrade strategy implementer.
type ControlPlaneExternalUpgradeResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	CommonRetryResponse `json:",inline"`
}

// MachinesExternalUpgradeRequest is the input to an external upgrade
// strategy implementer for a groups of machines.
type MachinesExternalUpgradeRequest struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRequest contains fields common to all request types.
	CommonRequest          `json:",inline"`
	Cluster                corev1.ObjectReference
	Owner                  corev1.ObjectReference
	MachinesRequireUpgrade []corev1.ObjectReference
	NewMachine             MachineSpec
}

// MachinesExternalUpgradeResponse is the response from an external
// upgrade strategy implementer.
type MachinesExternalUpgradeResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	CommonRetryResponse `json:",inline"`
}
