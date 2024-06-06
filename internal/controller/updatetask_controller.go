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

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
	"github.com/microsoft/cluster-api-inplace-updater/internal/controller/controlplanes"
	"github.com/microsoft/cluster-api-inplace-updater/pkg/nodes"
	"github.com/microsoft/cluster-api-inplace-updater/pkg/scopes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	waitRequeueAfter    = 30 * time.Second
	changedRequeueAfter = 100 * time.Millisecond
)

// UpdateTaskReconciler reconciles a UpdateTask object
type UpdateTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes;kubeadmcontrolplanes/status,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=dockernodeupdatetasktemplates,verbs=get;list;watch

//+kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=updatetasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=updatetasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=updatetasks/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=create;get;list;watch;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UpdateTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *UpdateTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (rret ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).WithValues("namespace", req.Namespace, "taskName", req.Name)

	updateTask := &updatev1beta1.UpdateTask{}
	if err := r.Get(ctx, req.NamespacedName, updateTask); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to get updateTask")
		return ctrl.Result{}, err
	}

	helper, err := patch.NewHelper(updateTask, r.Client)
	if err != nil {
		logger.Error(err, "unable to create task patcher")
		return ctrl.Result{}, err
	}

	defer func() {
		if updatev1beta1.IsTerminated(updateTask.Status.State) {
			controllerutil.RemoveFinalizer(updateTask, updatev1beta1.UpdateTaskFinalizer)
		} else if rerr == nil && rret.IsZero() {
			rret.Requeue = true
		}

		conditions.SetSummary(updateTask, conditions.WithConditions(
			updatev1beta1.UpdateOperationCondition,
			updatev1beta1.AbortOperationCondition,
			updatev1beta1.TrackOperationCondition,
		))

		err := helper.Patch(ctx, updateTask)
		if err != nil {
			logger.Error(err, "unable to patch update task")
			rerr = kerrors.NewAggregate([]error{rerr, err})
		}
	}()

	if !controllerutil.ContainsFinalizer(updateTask, updatev1beta1.UpdateTaskFinalizer) {
		controllerutil.AddFinalizer(updateTask, updatev1beta1.UpdateTaskFinalizer)
		// save immediately
		return ctrl.Result{RequeueAfter: changedRequeueAfter}, nil
	}

	scope, err := scopes.New(ctx, logger, r.Client, updateTask)
	if err != nil {
		return ctrl.Result{}, err
	}

	if result, err := r.reconcileStatus(ctx, logger, scope); err != nil {
		return result, err
	}

	if !updateTask.DeletionTimestamp.IsZero() || updateTask.Spec.TargetPhase == updatev1beta1.UpdateTaskPhaseAbort {
		if result, err := r.reconcileAbort(ctx, logger, scope); err != nil || !result.IsZero() {
			return result, err
		}
	} else if updateTask.Spec.TargetPhase == updatev1beta1.UpdateTaskPhaseUpdate {
		if result, err := r.reconcileNormal(ctx, logger, scope); err != nil || !result.IsZero() {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *UpdateTaskReconciler) reconcileAbort(ctx context.Context, logger logr.Logger, scope *scopes.UpdateTaskScope) (ctrl.Result, error) {
	conditions.Delete(scope.Task, updatev1beta1.UpdateOperationCondition)
	for _, nodeUpdateTask := range scope.NodeUpdateTasks {
		nodeUpdateStatus, err := nodes.GetNodeUpdateTaskStatus(ctx, r.Client, nodeUpdateTask)
		if err != nil {
			logger.Error(err, "cannot extract info from nodeUpdateTask",
				"apiVersion", nodeUpdateTask.GetAPIVersion(),
				"kind", nodeUpdateTask.GetKind(),
				"name", nodeUpdateTask.GetName())
			return ctrl.Result{}, err
		}

		if !updatev1beta1.IsTerminated(nodeUpdateStatus.State) {
			err := nodes.AbortNodeUpdateTask(ctx, r.Client, nodeUpdateTask)
			if err != nil {
				msg := "cannot abort nodeUpdateTask. " + err.Error()
				logger.Error(err, msg,
					"apiVersion", nodeUpdateTask.GetAPIVersion(),
					"kind", nodeUpdateTask.GetKind(),
					"name", nodeUpdateTask.GetName())
				conditions.MarkFalse(scope.Task, updatev1beta1.AbortOperationCondition, updatev1beta1.AbortNodeUpdateFailedReason, clusterv1.ConditionSeverityError, msg)
				return ctrl.Result{}, err
			}
		}
	}

	conditions.MarkTrue(scope.Task, updatev1beta1.AbortOperationCondition)
	return ctrl.Result{}, nil
}

func (r *UpdateTaskReconciler) reconcileNormal(ctx context.Context, logger logr.Logger, scope *scopes.UpdateTaskScope) (ctrl.Result, error) {
	nodeUpdateStatusMap := map[string]*nodes.NodeUpdateTaskStatus{}
	for _, nodeUpdateTask := range scope.NodeUpdateTasks {
		nodeUpdateStatus, err := nodes.GetNodeUpdateTaskStatus(ctx, r.Client, nodeUpdateTask)
		if err != nil {
			logger.Error(err, "cannot extract info from nodeUpdateTask",
				"apiVersion", nodeUpdateTask.GetAPIVersion(),
				"kind", nodeUpdateTask.GetKind(),
				"name", nodeUpdateTask.GetName())
			return ctrl.Result{}, err
		}

		mLogger := logger.WithValues("machine", nodeUpdateStatus.MachineName)
		if existingUpdateStatus, ok := nodeUpdateStatusMap[nodeUpdateStatus.MachineName]; ok {
			mLogger.Info("there are multiple nodeUpdateTasks for same machine, will pick lastest created one")
			if existingUpdateStatus.CreationTimestamp.Before(&nodeUpdateStatus.CreationTimestamp) {
				nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus
			}
		} else {
			nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus
		}

		if !updatev1beta1.IsTerminated(nodeUpdateStatus.State) {
			mLogger.Info("node update is inprogress")
			conditions.MarkTrue(scope.Task, updatev1beta1.UpdateOperationCondition)
			return ctrl.Result{}, nil
		} else if nodeUpdateStatus.State == updatev1beta1.UpdateTaskStateFailed {
			msg := "nodeUpdate failed, cluster UpdateTask will get blocked. If MachineHealthCheck is configured, then UpdateTask will continue once failed machine get remediated"
			mLogger.Info(msg)
			conditions.MarkFalse(scope.Task, updatev1beta1.UpdateOperationCondition, updatev1beta1.NodeUpdateFailedReason, clusterv1.ConditionSeverityError, msg)
			return ctrl.Result{}, nil
		}
	}

	//TODO: handle machineDeployment preflight check

	// handle controlplane preflight check
	if scope.Task.Spec.ControlPlaneRef != nil {
		if scope.Task.Spec.ControlPlaneRef.Kind == "KubeadmControlPlane" {
			if result, err := controlplanes.PreflightCheckForKubeadmControlPlane(ctx, logger, r.Client, scope.Task.Namespace, scope.Task.Spec.ControlPlaneRef); err != nil || !result.IsZero() {
				msg := "controlplane preflight check failed"
				logger.Info(msg)
				conditions.MarkFalse(scope.Task, updatev1beta1.UpdateOperationCondition, updatev1beta1.PreflightCheckFailedReason, clusterv1.ConditionSeverityWarning, msg)
				return result, err
			} else {
				conditions.MarkTrue(scope.Task, updatev1beta1.UpdateOperationCondition)
			}
		}
	}

	for _, machine := range scope.UpdateMachines.SortedByCreationTimestamp() {
		if _, ok := nodeUpdateStatusMap[machine.Name]; !ok {

			if !isMachineInUpdateScope(machine, scope.Task.Name) {
				continue
			}

			// found not updated machine!
			_, err := nodes.CreateNodeUpdateTask(ctx, r.Client, scope.Task, machine)
			if err != nil {
				msg := "unable to create nodeUpdateTask. " + err.Error()
				logger.Error(err, msg, "machine", machine.Name)
				conditions.MarkFalse(scope.Task, updatev1beta1.UpdateOperationCondition, updatev1beta1.CreateNodeUpdateFailedReason, clusterv1.ConditionSeverityError, msg)
				return ctrl.Result{}, err
			} else {
				logger.Info("created nodeUpdateTask", "machine", machine.Name)
			}

			conditions.MarkTrue(scope.Task, updatev1beta1.UpdateOperationCondition)
			return ctrl.Result{RequeueAfter: changedRequeueAfter}, nil
		}
	}

	conditions.MarkTrue(scope.Task, updatev1beta1.UpdateOperationCondition)
	return ctrl.Result{}, nil
}

func (r *UpdateTaskReconciler) reconcileStatus(ctx context.Context, logger logr.Logger, scope *scopes.UpdateTaskScope) (ctrl.Result, error) {
	nodeUpdateStatusMap := map[string]*nodes.NodeUpdateTaskStatus{}
	for _, nodeUpdateTask := range scope.NodeUpdateTasks {
		nodeUpdateStatus, err := nodes.GetNodeUpdateTaskStatus(ctx, r.Client, nodeUpdateTask)
		if err != nil {
			logger.Error(err, "cannot extract info from nodeUpdateTask",
				"apiVersion", nodeUpdateTask.GetAPIVersion(),
				"kind", nodeUpdateTask.GetKind(),
				"name", nodeUpdateTask.GetName())
			return ctrl.Result{}, err
		}

		mLogger := logger.WithValues("machine", nodeUpdateStatus.MachineName)
		if existingUpdateStatus, ok := nodeUpdateStatusMap[nodeUpdateStatus.MachineName]; ok {
			mLogger.Info("there are multiple nodeUpdateTasks for same machine, will pick lastest created one")
			if existingUpdateStatus.CreationTimestamp.Before(&nodeUpdateStatus.CreationTimestamp) {
				nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus
			}
		} else {
			nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus
		}
	}

	updated := 0
	aborted := 0
	failed := 0
	inProgress := 0
	waitForUpdate := 0
	total := 0
	machines := scope.UpdateMachines.SortedByCreationTimestamp()
	for _, machine := range machines {
		nodeUpdateStatus, ok := nodeUpdateStatusMap[machine.Name]
		if ok {
			switch nodeUpdateStatus.State {
			case updatev1beta1.UpdateTaskStateUpdated:
				updated++
			case updatev1beta1.UpdateTaskStateAborted:
				aborted++
			case updatev1beta1.UpdateTaskStateFailed:
				failed++
			case updatev1beta1.UpdateTaskStateInProgress, updatev1beta1.UpdateTaskStateUnknown:
				inProgress++
			}
			total++
		} else if isMachineInUpdateScope(machine, scope.Task.Name) {
			waitForUpdate++
			total++
		}

		// only update machine which in current scope
		if isMachineInUpdateScope(machine, scope.Task.Name) {
			machineHelper, err := patch.NewHelper(machine, r.Client)
			if err != nil {
				logger.Error(err, "unable to create patchHelper for machine", "machine", machine.Name)
				continue
			}

			if nodeUpdateStatus != nil {
				switch nodeUpdateStatus.State {
				case updatev1beta1.UpdateTaskStateUpdated:
					conditions.MarkTrue(machine, updatev1beta1.MachineUpToDate)
				case updatev1beta1.UpdateTaskStateAborted:
					conditions.MarkFalse(machine, updatev1beta1.MachineUpToDate, "Aborted", clusterv1.ConditionSeverityInfo, "Aborted")
				case updatev1beta1.UpdateTaskStateFailed:
					conditions.MarkFalse(machine, updatev1beta1.MachineUpToDate, "Failed", clusterv1.ConditionSeverityError, "Failed")
				case updatev1beta1.UpdateTaskStateInProgress:
					conditions.MarkFalse(machine, updatev1beta1.MachineUpToDate, "InProgress", clusterv1.ConditionSeverityInfo, "InProgress")
				}
			} else {
				conditions.MarkFalse(machine, updatev1beta1.MachineUpToDate, "WaitForUpdate", clusterv1.ConditionSeverityInfo, "WaitForUpdate")
			}

			err = machineHelper.Patch(ctx, machine)
			if err != nil {
				logger.Error(err, "unable to patch machine to update condition", "machine", machine.Name)
				continue
			}
		}
	}

	// TODO: updateTask conditions
	// aborting
	if !scope.Task.DeletionTimestamp.IsZero() || scope.Task.Spec.TargetPhase == updatev1beta1.UpdateTaskPhaseAbort {
		if inProgress == 0 {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateAborted
			conditions.MarkTrue(scope.Task, updatev1beta1.TrackOperationCondition)
			logger.Info("updateTask aborted")
		} else {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateInProgress
			msg := fmt.Sprintf("%d updated, %d failed, %d aborted, %d aborting, %d waitForUpdate", updated, failed, aborted, inProgress, waitForUpdate)
			conditions.MarkFalse(scope.Task, updatev1beta1.TrackOperationCondition, "Aborting", clusterv1.ConditionSeverityInfo, msg)
			logger.Info("updateTask aborting, " + msg)
		}
	} else { // updating
		if total == updated {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateUpdated
			conditions.MarkTrue(scope.Task, updatev1beta1.TrackOperationCondition)
			logger.Info("updateTask completed")
		} else {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateInProgress
			msg := fmt.Sprintf("%d updated, %d failed, %d aborted, %d inProgress, %d waitForUpdate", updated, failed, aborted, inProgress, waitForUpdate)
			conditions.MarkFalse(scope.Task, updatev1beta1.TrackOperationCondition, "InProgress", clusterv1.ConditionSeverityInfo, msg)
			logger.Info("updateTask inprogress, " + msg)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UpdateTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1beta1.UpdateTask{}).
		Complete(r)
}

func isMachineInUpdateScope(machine *clusterv1.Machine, taskName string) bool {
	runningTaskName, ok := machine.Annotations[updatev1beta1.UpdateTaskAnnotationName]
	return ok && runningTaskName == taskName
}
