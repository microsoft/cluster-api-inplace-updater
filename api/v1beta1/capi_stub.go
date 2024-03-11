package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
)

type MachineSpec struct {
	Version                         string                  `json:"version"`
	InfrastructureTemplateReference *corev1.ObjectReference `json:"infrastructureTemplateRef"`
}

// ControlPlaneExternalStrategy request external entity to make upgrade for controlplane
func ControlPlaneExternalStrategy(*ControlPlaneExternalStrategyRequest, *ExternalStrategyResponse) {}

// ControlPlaneExternalStrategy request external entity to make upgrade for machinedeployment
func MachineDeploymentExternalStrategy(*MachineDeploymentExternalStrategyRequest, *ExternalStrategyResponse) {
}

// From https://github.com/kubernetes-sigs/cluster-api/compare/main...g-gaston:cluster-api:in-place-upgrades-poc

// ExternalStrategyRequest is the input to an external upgrade
// strategy implementer.
type ControlPlaneExternalStrategyRequest struct {
	Cluster                *clusterv1.Cluster
	ControlPlane           *corev1.ObjectReference
	MachinesRequireUpgrade collections.Machines
	NewMachine             *MachineSpec
}

type MachineDeploymentExternalStrategyRequest struct {
	Cluster                *clusterv1.Cluster
	MachineDeployment      *corev1.ObjectReference
	MachinesRequireUpgrade collections.Machines
	NewMachine             *MachineSpec
}

// ExternalStrategyResponse is the response from an external
// upgrade strategy implementer.
type ExternalStrategyResponse struct {
	Accepted bool
	Reason   string
}
