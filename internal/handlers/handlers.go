package handlers

import (
	"context"

	upgradev1beta1 "github.com/microsoft/cluster-api-inplace-upgrader/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage/names"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExtensionHandlers struct {
	client client.Client
}

// NewExtensionHandlers returns a ExtensionHandlers for the lifecycle hooks handlers.
func NewExtensionHandlers(client client.Client) *ExtensionHandlers {
	return &ExtensionHandlers{
		client: client,
	}
}

func (m *ExtensionHandlers) DoBeforeClusterCreate(ctx context.Context, request *runtimehooksv1.BeforeClusterCreateRequest, response *runtimehooksv1.BeforeClusterCreateResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("BeforeClusterCreate is called")
}

// /home/azureuser/repo/cluster-api/controllers/external/util.go
func (m *ExtensionHandlers) HandleControlPlaneExternalStrategy(ctx context.Context, request *upgradev1beta1.ControlPlaneExternalStrategyRequest, response *upgradev1beta1.ExternalStrategyResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("ControlPlaneExternalStrategyRequest is called")

	// TODO: filter policy and pick 1
	polices := &upgradev1beta1.UpgradePolicyList{}
	if err := m.client.List(ctx, polices); err != nil {
		log.Error(err, "unable to list upgradePolicy")
		response.Accepted = false
		return
	}

	if len(polices.Items) == 0 {
		log.Info("no upgradePolicy found")
		response.Accepted = false
		return
	}

	policy := polices.Items[0]

	// create upgradeTask
	tasks := &upgradev1beta1.UpgradeTaskList{}
	if err := m.client.List(ctx, tasks); err != nil {
		log.Error(err, "unable to list upgradeTask")
		response.Accepted = false
		return
	}

	// abort existing upgrade if have
	for _, task := range tasks.Items {
		if task.Spec.Phase == upgradev1beta1.OngoingPhase &&
			task.Spec.ClusterRef.Namespace == request.Cluster.Namespace &&
			task.Spec.ClusterRef.Name == request.Cluster.Name &&
			*task.Spec.ControlPlaneRef == *request.ControlPlane {
			task.Spec.Phase = upgradev1beta1.AbortPhase
			abortPatch := client.RawPatch(types.MergePatchType, []byte("{\"spec\":{\"phase\":\"Abort\"}}"))
			if err := m.client.Patch(ctx, &task, abortPatch); err != nil {
				log.Error(err, "unable to patch existing upgradeTask")
				response.Accepted = false
				return
			}
		}
	}

	// create new upgradeTask

	upgradeMachines := []corev1.ObjectReference{}
	for _, m := range request.MachinesRequireUpgrade {
		upgradeMachines = append(upgradeMachines, corev1.ObjectReference{
			Kind:      m.Kind,
			Name:      m.Name,
			Namespace: m.Namespace,
		})
	}

	newName := names.SimpleNameGenerator.GenerateName(request.Cluster.Name + "-" + policy.Name + "-")
	newTask := &upgradev1beta1.UpgradeTask{
		TypeMeta: metav1.TypeMeta{
			APIVersion: upgradev1beta1.GroupVersion.String(),
			Kind:       "UpgradeTask",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: request.Cluster.Namespace,
			Name:      newName,
		},
		Spec: upgradev1beta1.UpgradeTaskSpec{
			ClusterRef: &corev1.ObjectReference{
				Kind:      request.Cluster.Kind,
				Name:      request.Cluster.Name,
				Namespace: request.Cluster.Namespace,
			},
			ControlPlaneRef:        request.ControlPlane.DeepCopy(),
			MachinesRequireUpgrade: upgradeMachines,
			NewMachineSpec:         request.NewMachine,
		},
	}

	newTask.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: policy.APIVersion,
			Kind:       policy.Kind,
			Name:       policy.Name,
			UID:        policy.UID,
		},
	})

	if err := m.client.Create(ctx, newTask); err != nil {
		log.Error(err, "unable to create upgradeTask")
		response.Accepted = false
		return
	}

	response.Accepted = true
}

func (m *ExtensionHandlers) HandleMachineDeploymentExternalStrategy(ctx context.Context, request *upgradev1beta1.MachineDeploymentExternalStrategyRequest, response *upgradev1beta1.ExternalStrategyResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("MachineDeploymentExternalStrategyRequest is called")

	response.Accepted = false
}
