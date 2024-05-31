package nodes

import (
	"context"

	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ClusterUpdateTaskNameLabel = "update.extension.cluster.x-k8s.io/cluster-update-task-name"

type NodeUpdateTaskStatus struct {
	CreationTimestamp metav1.Time                   `json:"creationTimestamp,omitempty"`
	MachineName       string                        `json:"machineName,omitempty"`
	Phase             updatev1beta1.UpdateTaskPhase `json:"phase,omitempty"`
	State             updatev1beta1.UpdateTaskState `json:"state,omitempty"`
	Conditions        clusterv1.Conditions          `json:"conditions,omitempty"`
}

func GetNodeUpdateTaskStatus(ctx context.Context, c client.Client, nodeUpdateTask *unstructured.Unstructured) (*NodeUpdateTaskStatus, error) {
	status := &NodeUpdateTaskStatus{}
	status.CreationTimestamp = nodeUpdateTask.GetCreationTimestamp()
	if err := util.UnstructuredUnmarshalField(nodeUpdateTask, &status.MachineName, "spec", "machineName"); err != nil {
		return nil, errors.Wrap(err, "unable to extract spec.machineName field from nodeupdatetask")
	}
	if err := util.UnstructuredUnmarshalField(nodeUpdateTask, &status.Phase, "spec", "phase"); err != nil {
		return nil, errors.Wrap(err, "unable to extract spec.phase field from nodeupdatetask")
	}
	if err := util.UnstructuredUnmarshalField(nodeUpdateTask, &status.State, "status", "state"); err != nil {
		// it's possible status not initialized
		status.State = updatev1beta1.UpdateTaskStateUnknown
	}
	_ = util.UnstructuredUnmarshalField(nodeUpdateTask, &status.Conditions, "status", "conditions")

	return status, nil
}

func AbortNodeUpdateTask(ctx context.Context, c client.Client, nodeUpdateTask *unstructured.Unstructured) error {
	patchData := []byte(`{"spec": {"phase": "Abort"}}`)
	return c.Patch(ctx, nodeUpdateTask, client.RawPatch(types.MergePatchType, patchData))
}

// Create nodeUpdateTask from template.
// nodeUpdateTask must has a ClusterUpdateTaskNameLabel label point to cluster updateTask.
// nodeUpdateTask must has spec.phase with type UpdateTaskPhase.\
// nodeUpdateTask must has spec.newMachineSpec with type MachineSpec.
// nodeUpdateTask must has status.state with type UpdateTaskState.
// nodeUpdateTask can has status.conditions to report detailed state message.
func CreateNodeUpdateTask(ctx context.Context, c client.Client, task *updatev1beta1.UpdateTask, machine *clusterv1.Machine) (*unstructured.Unstructured, error) {
	template, err := external.Get(ctx, c, &task.Spec.NodeUpdateTemplate.InfrastructureRef, task.Namespace)
	if err != nil {
		return nil, err
	}

	labels := task.Spec.NodeUpdateTemplate.Labels
	labels[clusterv1.ClusterNameLabel] = task.Spec.ClusterRef.Name
	labels[ClusterUpdateTaskNameLabel] = task.Name

	taskRef := &metav1.OwnerReference{
		Kind:       task.Kind,
		APIVersion: task.APIVersion,
		Name:       task.Name,
		UID:        task.UID,
		Controller: ptr.To(true),
	}

	nodeUpdateTask, err := external.GenerateTemplate(&external.GenerateTemplateInput{
		Template:    template,
		TemplateRef: &task.Spec.NodeUpdateTemplate.InfrastructureRef,
		Namespace:   task.Namespace,
		Labels:      labels,
		Annotations: task.Spec.NodeUpdateTemplate.Annotations,
		OwnerRef:    taskRef,
	})

	if err != nil {
		return nil, err
	}

	if err := unstructured.SetNestedField(nodeUpdateTask.Object, task.Spec.NewMachineSpec.BootstrapConfig, "spec", "newMachineSpec", "bootstrapConfig"); err != nil {
		return nil, errors.Wrap(err, "unable to set spec.newMachineSpec.bootstrapConfig for nodeupdatetask")
	}

	if err := unstructured.SetNestedField(nodeUpdateTask.Object, task.Spec.NewMachineSpec.InfraMachine, "spec", "newMachineSpec", "infraMachine"); err != nil {
		return nil, errors.Wrap(err, "unable to set spec.newMachineSpec.infraMachine for nodeupdatetask")
	}

	if err := unstructured.SetNestedField(nodeUpdateTask.Object, machine.Name, "spec", "machineName"); err != nil {
		return nil, errors.Wrap(err, "unable to set spec.machineName for nodeupdatetask")
	}

	err = c.Create(ctx, nodeUpdateTask)
	if err != nil {
		err = errors.Wrap(err, "create nodeupdatetask failed")
	}

	return nodeUpdateTask, err
}
