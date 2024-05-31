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
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
)

// TODO: how to define machine spec???
/**
// kubeadmControlPlane
   kubeadmConfigSpec:
      clusterConfiguration:
        apiServer:
          certSANs:
          - localhost
          - 127.0.0.1
          - 0.0.0.0
          - host.docker.internal
        controllerManager:
          extraArgs:
            enable-hostpath-provisioner: "true"
        dns: {}
        etcd: {}
        networking: {}
        scheduler: {}
      format: cloud-config
      initConfiguration:
        localAPIEndpoint: {}
        nodeRegistration:
          imagePullPolicy: IfNotPresent
      joinConfiguration:
        discovery: {}
        nodeRegistration:
          imagePullPolicy: IfNotPresent
    machineTemplate:
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
        kind: DockerMachineTemplate
        name: md-scale-au4y6s-control-plane
        namespace: md-scale-c12jos
      metadata: {}

// machineDeployment
    template:
      metadata:
        labels:
          cluster.x-k8s.io/cluster-name: md-scale-au4y6s
          cluster.x-k8s.io/deployment-name: md-scale-au4y6s-md-0
      spec:
        bootstrap:
          configRef:
            apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
            kind: KubeadmConfigTemplate
            name: md-scale-au4y6s-md-0
        clusterName: md-scale-au4y6s
        failureDomain: fd4
        infrastructureRef:
          apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
          kind: DockerMachineTemplate
          name: md-scale-au4y6s-md-0
        version: v1.29.0
**/
type MachineSpec struct {
	BootstrapConfig string `json:"bootstrapConfig,omitempty"`
	InfraMachine    string `json:"infraMachine,omitempty"`
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

const MachineUpToDate clusterv1.ConditionType = "MachineUpToDate"
