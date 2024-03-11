package v1beta1

import corev1 "k8s.io/api/core/v1"

type MachineSpec struct {
	Version                         string                  `json:"version"`
	InfrastructureTemplateReference *corev1.ObjectReference `json:"infrastructureTemplateRef"`
}
