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
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/test/infrastructure/container"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
)

// DockerNodeUpdateTaskReconciler reconciles a DockerNodeUpdateTask object
type DockerNodeUpdateTaskReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=dockernodeupdatetasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=dockernodeupdatetasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=update.extension.cluster.x-k8s.io,resources=dockernodeupdatetasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DockerNodeUpdateTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *DockerNodeUpdateTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	task := &updatev1beta1.DockerNodeUpdateTask{}
	if err := r.Get(ctx, req.NamespacedName, task); err != nil {
		logger.Error(err, "failed to get task")
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	helper, err := patch.NewHelper(task, r.Client)
	if err != nil {
		logger.Error(err, "failed to create task patcher")
		return ctrl.Result{}, err
	}

	defer func() {
		err := helper.Patch(ctx, task)
		if err != nil {
			logger.Error(err, "fail to patch task")
		}
	}()

	machine := &clusterv1.Machine{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: task.Spec.MachineName}, machine); err != nil {
		logger.Error(err, "failed to get machine")
		return ctrl.Result{}, err
	}

	// dockerMachine := &infrav1.DockerMachine{}
	// if err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: machine.Spec.InfrastructureRef.Name}, machine); err != nil {
	// 	logger.Error(err, "failed to get dockermachine")
	// 	return ctrl.Result{}, err
	// }

	runtime, err := container.NewDockerClient()
	if err != nil {
		logger.Error(err, "failed to create docker runtime")
		return ctrl.Result{}, err
	}

	if task.Spec.Phase == updatev1beta1.UpdateTaskPhaseUpdate && !updatev1beta1.IsTerminated(task.Status.State) {
		var outErr bytes.Buffer
		var outStd bytes.Buffer
		execInput := &container.ExecContainerInput{
			OutputBuffer: &outStd,
			ErrorBuffer:  &outErr,
			InputBuffer:  strings.NewReader("echo 111 " + task.Spec.MachineName),
		}

		if err := runtime.ExecContainer(ctx, machine.Name, execInput, "echo 222 "+task.Spec.MachineName); err != nil {
			task.Status.State = updatev1beta1.UpdateTaskStateFailed

			// TODO: fix permission issue
			// err="error creating container exec: permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Post \"http://%2Fvar%2Frun%2Fdocker.sock/v1.24/containers/k8s-inplace-upgrade-j29fdn-x4286-r8rcb/exec\": dial unix /var/run/docker.sock: connect: permission denied" controller="dockernodeupdatetask"
			logger.Error(err, "fail to exec")
			//return ctrl.Result{}, err
		}
		time.Sleep(10 * time.Second)

		task.Status.State = updatev1beta1.UpdateTaskStateUpdated
		logger.Info(fmt.Sprintf("command exec std:%s, err:%s", outStd.String(), outErr.String()))
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DockerNodeUpdateTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&updatev1beta1.DockerNodeUpdateTask{}).
		Complete(r)
}
