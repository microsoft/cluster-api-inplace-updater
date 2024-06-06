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

package lifecycle

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UnexpectedError = "UnexpectedError"
	ClusterDeleting = "ClusterDeleting"
)

type ExtensionHandlers struct {
	client client.Client
}

func NewExtensionHandlers(client client.Client) *ExtensionHandlers {
	return &ExtensionHandlers{
		client: client,
	}
}

// TODO: move failable operations into CR controller, so that i can retryable.
func (m *ExtensionHandlers) DoControlPlaneExternalUpdate(ctx context.Context, request *updatev1beta1.ControlPlaneExternalUpdateRequest, response *updatev1beta1.ControlPlaneExternalUpdateResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("ControlPlaneExternalUpdate is called")

	err := m.doControlPlaneExternalUpdateImpl(ctx, request)
	if err != nil {
		response.SetStatus(runtimehooksv1.ResponseStatusFailure)
		response.SetMessage(err.Error())
		log.Error(err, "external update webhook rejected")
	} else {
		response.SetStatus(runtimehooksv1.ResponseStatusSuccess)
		log.Info("external update webhook accepted")
	}
}

func (m *ExtensionHandlers) doControlPlaneExternalUpdateImpl(ctx context.Context, request *updatev1beta1.ControlPlaneExternalUpdateRequest) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("ControlPlaneExternalUpdate is called")

	if request.ClusterRef == nil {
		return errors.New("clusterRef is nil")
	}

	cluster := &clusterv1.Cluster{}
	if err := m.client.Get(ctx, types.NamespacedName{Namespace: request.ClusterRef.Namespace, Name: request.ClusterRef.Name}, cluster); err != nil {
		return errors.Wrap(err, "get cluster failed")
	}

	if cluster.DeletionTimestamp != nil {
		return errors.New("cluster deleting")
	}

	if request.ControlPlaneRef == nil {
		return errors.New("controlPlaneRef is nil")
	}

	updateTaskList := &updatev1beta1.UpdateTaskList{}
	if err := m.client.List(ctx, updateTaskList, client.InNamespace(request.ClusterRef.Namespace), client.MatchingLabels{clusterv1.ClusterNameLabel: request.ClusterRef.Name}); err != nil {
		return errors.Wrap(err, "list updateTask failed")
	}

	for _, task := range updateTaskList.Items {
		if task.Spec.ClusterRef.Name == request.ClusterRef.Name &&
			task.Spec.ControlPlaneRef.Name == request.ControlPlaneRef.Name &&
			task.Status.State == updatev1beta1.UpdateTaskStateInProgress {

			patch := client.RawPatch(types.MergePatchType, []byte(`{"spec":{"phase":"Abort"}}`))
			if err := m.client.Patch(ctx, &task, patch); err != nil {
				return errors.Wrap(err, "abort existing updateTask failed")
			}
		}
	}

	newTaskName := fmt.Sprintf("%s-update-%s", request.ClusterRef.Name, util.RandomString(6))
	machineList := &clusterv1.MachineList{}
	if err := m.client.List(ctx, machineList, client.InNamespace(request.ClusterRef.Namespace), client.MatchingLabels{clusterv1.ClusterNameLabel: request.ClusterRef.Name, clusterv1.MachineControlPlaneLabel: ""}); err != nil {
		return errors.Wrap(err, "list machine failed")
	}

	for _, machine := range machineList.Items {
		patchHelper, err := patch.NewHelper(&machine, m.client)
		if err != nil {
			return errors.Wrap(err, "create machine patcher failed")
		}

		machine.Annotations[updatev1beta1.UpdateTaskAnnotationName] = newTaskName
		conditions.Delete(&machine, updatev1beta1.MachineUpToDate)

		if err != patchHelper.Patch(ctx, &machine) {
			return errors.Wrap(err, "patch machine failed")
		}
	}

	policyList := &updatev1beta1.UpdatePolicyList{}
	if err := m.client.List(ctx, policyList, client.InNamespace(request.ClusterRef.Namespace)); err != nil {
		return errors.Wrap(err, "cannot find update policy")
	}

	if len(policyList.Items) == 0 {
		return errors.New("no policy found")
	}

	//TODO: filter policy
	policy := &policyList.Items[0]
	newTask := &updatev1beta1.UpdateTask{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: request.ClusterRef.Namespace,
			Name:      newTaskName,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: request.ClusterRef.Name,
			},
			Annotations: map[string]string{},
		},
		Spec: updatev1beta1.UpdateTaskSpec{
			ClusterRef:            request.ClusterRef.DeepCopy(),
			ControlPlaneRef:       request.ControlPlaneRef.DeepCopy(),
			MachinesRequireUpdate: []corev1.ObjectReference{},
			NewMachineSpec:        *request.NewMachine.DeepCopy(),
			NodeUpdateTemplate: updatev1beta1.NodeUpdateTaskTemplateSpec{
				InfrastructureRef: *policy.Spec.NodeUpdateTemplateRef.DeepCopy(),
			},
			TargetPhase: updatev1beta1.UpdateTaskPhaseUpdate,
		},
	}
	if err := m.client.Create(ctx, newTask); err != nil {
		return errors.Wrap(err, "unable to create new updateTask")
	}

	return nil
}

func (m *ExtensionHandlers) DoMachineDeploymentExternalUpdate(ctx context.Context, request *updatev1beta1.MachineDeploymentExternalUpdateRequest, response *updatev1beta1.MachineDeploymentExternalUpdateResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("DoMachineDeploymentExternalUpdate is called")

	//TODO: implement
}

// DoBeforeClusterUpdate implements the HandlerFunc for the BeforeClusterUpdate hook.
func (m *ExtensionHandlers) DoBeforeClusterUpgrade(ctx context.Context, request *runtimehooksv1.BeforeClusterUpgradeRequest, response *runtimehooksv1.BeforeClusterUpgradeResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("BeforeClusterUpgrade is called")
}

// DoAfterControlPlaneUpgrade implements the HandlerFunc for the AfterControlPlaneUpgrade hook.
func (m *ExtensionHandlers) DoAfterControlPlaneUpgrade(ctx context.Context, request *runtimehooksv1.AfterControlPlaneUpgradeRequest, response *runtimehooksv1.AfterControlPlaneUpgradeResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("AfterControlPlaneUpgrade is called")
}

// DoAfterClusterUpgrade implements the HandlerFunc for the AfterClusterUpgrade hook.
func (m *ExtensionHandlers) DoAfterClusterUpgrade(ctx context.Context, request *runtimehooksv1.AfterClusterUpgradeRequest, response *runtimehooksv1.AfterClusterUpgradeResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("AfterClusterUpgrade is called")
}

func (m *ExtensionHandlers) DoBeforeClusterCreate(ctx context.Context, request *runtimehooksv1.BeforeClusterCreateRequest, response *runtimehooksv1.BeforeClusterCreateResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("BeforeClusterCreate is called")
}
