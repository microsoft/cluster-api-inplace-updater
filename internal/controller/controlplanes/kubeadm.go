package controlplanes

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// deleteRequeueAfter is how long to wait before checking again to see if
	// all control plane machines have been deleted.
	deleteRequeueAfter = 30 * time.Second

	// preflightFailedRequeueAfter is how long to wait before trying to scale
	// up/down if some preflight check for those operation has failed.
	preflightFailedRequeueAfter = 15 * time.Second

	// dependentCertRequeueAfter is how long to wait before checking again to see if
	// dependent certificates have been created.
	dependentCertRequeueAfter = 30 * time.Second
)

func PreflightCheckForKubeadmControlPlane(ctx context.Context, logger logr.Logger, c client.Client, controlPlaneRef *corev1.ObjectReference) (ctrl.Result, error) {
	controlPlane := &controlplanev1.KubeadmControlPlane{}
	err := c.Get(ctx, types.NamespacedName{Namespace: controlPlaneRef.Namespace, Name: controlPlaneRef.Name}, controlPlane)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "cannot get KubeadmControlPlane from reference")
	}

	return preflightChecks(ctx, logger, c, controlPlane)
}

// preflightChecks checks if the control plane is stable before proceeding with a scale up/scale down operation,
// where stable means that:
// - There are no machine deletion in progress
// - All the health conditions on KCP are true.
// - All the health conditions on the control plane machines are true.
// If the control plane is not passing preflight checks, it requeue.
//
// NOTE: this func uses KCP conditions, it is required to call reconcileControlPlaneConditions before this.
func preflightChecks(ctx context.Context, logger logr.Logger, c client.Client, controlPlane *controlplanev1.KubeadmControlPlane, excludeFor ...*clusterv1.Machine) (ctrl.Result, error) { //nolint:unparam

	machineList := &clusterv1.MachineList{}
	clusterName, ok := controlPlane.Labels[clusterv1.ClusterNameLabel]
	if !ok {
		return ctrl.Result{}, errors.New("controlplane don't have clusterName label")
	}

	err := c.List(ctx, machineList, client.InNamespace(controlPlane.Namespace), client.MatchingLabels{
		clusterv1.ClusterNameLabel:             clusterName,
		clusterv1.MachineControlPlaneNameLabel: controlPlane.Name,
	})
	if err != nil {
		logger.Error(err, "unable to find controlplane machines", "clusterName", clusterName)
		return ctrl.Result{}, err
	}
	cpMachines := collections.FromMachineList(machineList)

	// If there is no KCP-owned control-plane machines, then control-plane has not been initialized yet,
	// so it is considered ok to proceed.
	if cpMachines.Len() == 0 {
		return ctrl.Result{}, nil
	}

	// If there are deleting machines, wait for the operation to complete.
	if len(cpMachines.Filter(collections.HasDeletionTimestamp)) > 0 {
		logger.Info("Waiting for machines to be deleted", "Machines", strings.Join(cpMachines.Filter(collections.HasDeletionTimestamp).Names(), ", "))
		return ctrl.Result{RequeueAfter: deleteRequeueAfter}, nil
	}

	// Check machine health conditions; if there are conditions with False or Unknown, then wait.
	allMachineHealthConditions := []clusterv1.ConditionType{
		controlplanev1.MachineAPIServerPodHealthyCondition,
		controlplanev1.MachineControllerManagerPodHealthyCondition,
		controlplanev1.MachineSchedulerPodHealthyCondition,
	}
	if isEtcdManaged(controlPlane) {
		allMachineHealthConditions = append(allMachineHealthConditions,
			controlplanev1.MachineEtcdPodHealthyCondition,
			controlplanev1.MachineEtcdMemberHealthyCondition,
		)
	}
	machineErrors := []error{}

loopmachines:
	for _, machine := range cpMachines {
		for _, excluded := range excludeFor {
			// If this machine should be excluded from the individual
			// health check, continue the out loop.
			if machine.Name == excluded.Name {
				continue loopmachines
			}
		}

		if machine.Status.NodeRef == nil {
			// The conditions will only ever be set on a Machine if we're able to correlate a Machine to a Node.
			// Correlating Machines to Nodes requires the nodeRef to be set.
			// Instead of confusing users with errors about that the conditions are not set, let's point them
			// towards the unset nodeRef (which is the root cause of the conditions not being there).
			machineErrors = append(machineErrors, errors.Errorf("Machine %s does not have a corresponding Node yet (Machine.status.nodeRef not set)", machine.Name))
		} else {
			for _, condition := range allMachineHealthConditions {
				if err := preflightCheckCondition("Machine", machine, condition); err != nil {
					machineErrors = append(machineErrors, err)
				}
			}
		}
	}
	if len(machineErrors) > 0 {
		aggregatedError := kerrors.NewAggregate(machineErrors)
		logger.Info("Waiting for control plane to pass preflight checks", "failures", aggregatedError.Error())

		return ctrl.Result{RequeueAfter: preflightFailedRequeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

func isEtcdManaged(controlPlane *controlplanev1.KubeadmControlPlane) bool {
	return controlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration == nil || controlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External == nil
}

func preflightCheckCondition(kind string, obj conditions.Getter, condition clusterv1.ConditionType) error {
	c := conditions.Get(obj, condition)
	if c == nil {
		return errors.Errorf("%s %s does not have %s condition", kind, obj.GetName(), condition)
	}
	if c.Status == corev1.ConditionFalse {
		return errors.Errorf("%s %s reports %s condition is false (%s, %s)", kind, obj.GetName(), condition, c.Severity, c.Message)
	}
	if c.Status == corev1.ConditionUnknown {
		return errors.Errorf("%s %s reports %s condition is unknown (%s)", kind, obj.GetName(), condition, c.Message)
	}
	return nil
}
