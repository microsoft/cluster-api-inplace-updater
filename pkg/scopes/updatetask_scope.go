package scopes

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
	"github.com/microsoft/cluster-api-inplace-updater/pkg/nodes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/errors"
)

type UpdateTaskScope struct {
	Task            *updatev1beta1.UpdateTask
	Cluster         *clusterv1.Cluster
	UpdateMachines  collections.Machines
	NodeUpdateTasks []*unstructured.Unstructured

	c           client.Client
	patchHelper *patch.Helper
}

func New(ctx context.Context, logger logr.Logger, c client.Client, task *updatev1beta1.UpdateTask) (*UpdateTaskScope, error) {
	namespace := task.Namespace
	cluster := &clusterv1.Cluster{}
	err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: task.Spec.ClusterRef.Name}, cluster)
	if err != nil {
		logger.Error(err, "unable to find cluster", "clusterName", task.Spec.ClusterRef.Name)
		return nil, err
	}

	var machines collections.Machines
	if task.Spec.ControlPlaneRef != nil {
		machineList := &clusterv1.MachineList{}
		err := c.List(ctx, machineList, client.InNamespace(namespace), client.MatchingLabels{
			clusterv1.ClusterNameLabel:         cluster.Name,
			clusterv1.MachineControlPlaneLabel: "",
		})
		if err != nil {
			logger.Error(err, "unable to find controlplane machines", "clusterName", task.Spec.ClusterRef.Name)
			return nil, err
		}
		machines = collections.FromMachineList(machineList)
	} else if task.Spec.MachineDeploymentRef != nil {
		machineList := &clusterv1.MachineList{}
		err := c.List(ctx, machineList, client.InNamespace(namespace), client.MatchingLabels{
			clusterv1.ClusterNameLabel:           cluster.Name,
			clusterv1.MachineDeploymentNameLabel: task.Spec.MachineDeploymentRef.Name,
		})
		if err != nil {
			logger.Error(err, "unable to find machinedeployment machines", "clusterName", task.Spec.ClusterRef.Name, "machinedeployment", task.Spec.MachineDeploymentRef.Name)
			return nil, err
		}
		machines = collections.FromMachineList(machineList)
	} else {
		err := errors.New("invalid updateTask spec, both controlplane and machinedeployment ref is nil")
		logger.Error(err, err.Error())
		return nil, err
	}

	nodeUpdateList := &unstructured.UnstructuredList{}
	nodeUpdateList.SetAPIVersion(task.Spec.NodeUpdateTemplate.InfrastructureRef.APIVersion)
	nodeUpdateList.SetKind(strings.TrimSuffix(task.Spec.NodeUpdateTemplate.InfrastructureRef.Kind, "Template"))
	if err := c.List(ctx, nodeUpdateList, client.InNamespace(namespace), client.MatchingLabels{nodes.ClusterUpdateTaskNameLabel: task.Name}); err != nil {
		logger.Error(err, "unable to list nodeUpdate CRs", "clusterName", task.Spec.ClusterRef.Name)
		return nil, err
	}
	nodeUpdateTasks := []*unstructured.Unstructured{}
	for _, nodeUpdateTask := range nodeUpdateList.Items {
		nodeUpdateTasks = append(nodeUpdateTasks, nodeUpdateTask.DeepCopy())
	}

	patchHelper, err := patch.NewHelper(task, c)
	if err != nil {
		logger.Error(err, "unable to create task patcher", "clusterName", task.Spec.ClusterRef.Name)
		return nil, err
	}

	return &UpdateTaskScope{
		Task:            task,
		Cluster:         cluster,
		UpdateMachines:  machines,
		NodeUpdateTasks: nodeUpdateTasks,
		patchHelper:     patchHelper,
		c:               c,
	}, nil
}

func (s *UpdateTaskScope) IsControlPlane() bool {
	return s.Task.Spec.ControlPlaneRef != nil
}
