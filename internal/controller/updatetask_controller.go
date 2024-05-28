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

	"github.com/go-logr/logr"
	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
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

// UpdateTaskReconciler reconciles a UpdateTask object
type UpdateTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
		} else if rerr == nil {
			rret.Requeue = true
		}

		err := helper.Patch(ctx, updateTask)
		if err != nil {
			logger.Error(err, "unable to patch update task")
			rerr = kerrors.NewAggregate([]error{rerr, err})
		}
	}()

	if !controllerutil.ContainsFinalizer(updateTask, updatev1beta1.UpdateTaskFinalizer) {
		controllerutil.AddFinalizer(updateTask, updatev1beta1.UpdateTaskFinalizer)
		// save immediately
		return ctrl.Result{Requeue: true}, nil
	}

	scope, err := scopes.New(ctx, logger, r.Client, updateTask)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !updateTask.DeletionTimestamp.IsZero() || updateTask.Spec.TargetPhase == updatev1beta1.UpdateTaskPhaseAbort {
		if result, err := r.reconcileAbort(ctx, logger, scope); err != nil {
			return result, err
		}
	} else if updateTask.Spec.TargetPhase == updatev1beta1.UpdateTaskPhaseUpdate {
		if result, err := r.reconcileNormal(ctx, logger, scope); err != nil {
			return result, err
		}
	}

	if result, err := r.reconcileStatus(ctx, logger, scope); err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *UpdateTaskReconciler) reconcileAbort(ctx context.Context, logger logr.Logger, scope *scopes.UpdateTaskScope) (ctrl.Result, error) {
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
				logger.Error(err, "cannot abort nodeUpdateTask",
					"apiVersion", nodeUpdateTask.GetAPIVersion(),
					"kind", nodeUpdateTask.GetKind(),
					"name", nodeUpdateTask.GetName())
				return ctrl.Result{}, err
			}
		}
	}

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
		nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus

		if !updatev1beta1.IsTerminated(nodeUpdateStatus.State) {
			logger.Info("node update is inprogress", "machine", nodeUpdateStatus.MachineName)
			return ctrl.Result{}, nil
		}
	}

	// TODO: handle controlplane & preflight check
	for _, machine := range scope.UpdateMachines.SortedByCreationTimestamp() {
		if _, ok := nodeUpdateStatusMap[machine.Name]; !ok {

			// for new created machine, or new updateTask take in control, just skip it.
			runningTaskName, ok := machine.Annotations[updatev1beta1.UpdateTaskAnnotationName]
			if !ok || runningTaskName != scope.Task.Name {
				continue
			}

			// found not updated machine!
			_, err := nodes.CreateNodeUpdateTask(ctx, r.Client, scope.Task, machine)
			if err != nil {
				logger.Error(err, "unable to create nodeUpdateTask", "machine", machine.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("created nodeUpdateTask", "machine", machine.Name)
			}

			return ctrl.Result{}, nil
		}
	}

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
		nodeUpdateStatusMap[nodeUpdateStatus.MachineName] = nodeUpdateStatus
	}

	updated := 0
	inProgress := 0
	waitForUpdate := 0
	total := 0
	machines := scope.UpdateMachines.SortedByCreationTimestamp()
	for _, machine := range machines {
		if nodeUpdateStatus, ok := nodeUpdateStatusMap[machine.Name]; ok {
			if updatev1beta1.IsTerminated(nodeUpdateStatus.State) {
				updated++
			} else {
				inProgress++
			}
		} else {
			waitForUpdate++
		}

		runningTaskName, ok := machine.Annotations[updatev1beta1.UpdateTaskAnnotationName]
		if ok && runningTaskName == scope.Task.Name {
			total++
			machineHelper, err := patch.NewHelper(machine, r.Client)
			if err != nil {
				logger.Error(err, "unable to create patchHelper for machine", "machine", machine.Name)
				continue
			}

			if nodeUpdateStatus, ok := nodeUpdateStatusMap[machine.Name]; ok {
				if updatev1beta1.IsTerminated(nodeUpdateStatus.State) {
					if nodeUpdateStatus.Phase == updatev1beta1.UpdateTaskPhaseAbort {
						conditions.MarkFalse(machine, updatev1beta1.MachineUpToDate, "Aborted", clusterv1.ConditionSeverityInfo, "Aborted")
					} else {
						conditions.MarkTrue(machine, updatev1beta1.MachineUpToDate)
					}
				} else {
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
			conditions.MarkTrue(scope.Task, clusterv1.ReadyCondition)
			logger.Info("updateTask aborted")
		} else {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateInProgress
			msg := fmt.Sprintf("%d updated, %d aborting, %d waitForUpdate", updated, inProgress, waitForUpdate)
			conditions.MarkFalse(scope.Task, clusterv1.ReadyCondition, "Aborting", clusterv1.ConditionSeverityInfo, msg)
			logger.Info("updateTask aborting, " + msg)
		}
	} else { // updating
		if total == updated {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateUpdated
			conditions.MarkTrue(scope.Task, clusterv1.ReadyCondition)
			logger.Info("updateTask completed")
		} else {
			scope.Task.Status.State = updatev1beta1.UpdateTaskStateInProgress
			msg := fmt.Sprintf("%d updated, %d inProgress, %d waitForUpdate", updated, inProgress, waitForUpdate)
			conditions.MarkFalse(scope.Task, clusterv1.ReadyCondition, "InProgress", clusterv1.ConditionSeverityInfo, msg)
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
