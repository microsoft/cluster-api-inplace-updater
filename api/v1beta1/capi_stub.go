package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineSpec struct {
	Machine         *clusterv1.Machine         `json:"machine,omitempty"`
	BootstrapConfig *unstructured.Unstructured `json:"bootstrapConfig,omitempty"`
	InfraMachine    *unstructured.Unstructured `json:"infraMachine,omitempty"`
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
	ClusterRef             *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	ControlPlaneRef        *corev1.ObjectReference  `json:"controlPlaneRef,omitempty"`
	MachinesRequireUpgrade []corev1.ObjectReference `json:"machinesRequireUpgrade,omitempty"`
	NewMachine             MachineSpec              `json:"newMachine,omitempty"`
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
	ClusterRef             *corev1.ObjectReference  `json:"clusterRef,omitempty"`
	MachineDeploymentRef   *corev1.ObjectReference  `json:"machineDeploymentRef,omitempty"`
	MachinesRequireUpgrade []corev1.ObjectReference `json:"machinesRequireUpgrade,omitempty"`
	NewMachine             MachineSpec  			`json:"newMachine,omitempty"`
}

// MachinesExternalUpgradeResponse is the response from an external
// upgrade strategy implementer.
type MachinesExternalUpgradeResponse struct {
	metav1.TypeMeta `json:",inline"`
	// CommonRetryResponse contains Status, Message and RetryAfterSeconds fields.
	CommonRetryResponse `json:",inline"`
}
