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
	"k8s.io/api/core/v1"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneExternalUpgradeRequest) DeepCopyInto(out *ControlPlaneExternalUpgradeRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.CommonRequest.DeepCopyInto(&out.CommonRequest)
	if in.ClusterRef != nil {
		in, out := &in.ClusterRef, &out.ClusterRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.ControlPlaneRef != nil {
		in, out := &in.ControlPlaneRef, &out.ControlPlaneRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.MachinesRequireUpgrade != nil {
		in, out := &in.MachinesRequireUpgrade, &out.MachinesRequireUpgrade
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	in.NewMachine.DeepCopyInto(&out.NewMachine)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneExternalUpgradeRequest.
func (in *ControlPlaneExternalUpgradeRequest) DeepCopy() *ControlPlaneExternalUpgradeRequest {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneExternalUpgradeRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneExternalUpgradeResponse) DeepCopyInto(out *ControlPlaneExternalUpgradeResponse) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.CommonRetryResponse = in.CommonRetryResponse
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneExternalUpgradeResponse.
func (in *ControlPlaneExternalUpgradeResponse) DeepCopy() *ControlPlaneExternalUpgradeResponse {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneExternalUpgradeResponse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineDeploymentExternalUpgradeRequest) DeepCopyInto(out *MachineDeploymentExternalUpgradeRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.CommonRequest.DeepCopyInto(&out.CommonRequest)
	if in.ClusterRef != nil {
		in, out := &in.ClusterRef, &out.ClusterRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.MachineDeploymentRef != nil {
		in, out := &in.MachineDeploymentRef, &out.MachineDeploymentRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.MachinesRequireUpgrade != nil {
		in, out := &in.MachinesRequireUpgrade, &out.MachinesRequireUpgrade
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	in.NewMachine.DeepCopyInto(&out.NewMachine)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineDeploymentExternalUpgradeRequest.
func (in *MachineDeploymentExternalUpgradeRequest) DeepCopy() *MachineDeploymentExternalUpgradeRequest {
	if in == nil {
		return nil
	}
	out := new(MachineDeploymentExternalUpgradeRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineDeploymentExternalUpgradeResponse) DeepCopyInto(out *MachineDeploymentExternalUpgradeResponse) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.CommonRetryResponse = in.CommonRetryResponse
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineDeploymentExternalUpgradeResponse.
func (in *MachineDeploymentExternalUpgradeResponse) DeepCopy() *MachineDeploymentExternalUpgradeResponse {
	if in == nil {
		return nil
	}
	out := new(MachineDeploymentExternalUpgradeResponse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineSpec) DeepCopyInto(out *MachineSpec) {
	*out = *in
	if in.Machine != nil {
		in, out := &in.Machine, &out.Machine
		*out = new(clusterv1.Machine)
		(*in).DeepCopyInto(*out)
	}
	if in.BootstrapConfig != nil {
		in, out := &in.BootstrapConfig, &out.BootstrapConfig
		*out = (*in).DeepCopy()
	}
	if in.InfraMachine != nil {
		in, out := &in.InfraMachine, &out.InfraMachine
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineSpec.
func (in *MachineSpec) DeepCopy() *MachineSpec {
	if in == nil {
		return nil
	}
	out := new(MachineSpec)
	in.DeepCopyInto(out)
	return out
}