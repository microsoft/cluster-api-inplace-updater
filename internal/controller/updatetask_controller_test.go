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
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	updatev1beta1 "github.com/microsoft/cluster-api-inplace-updater/api/v1beta1"
	"github.com/microsoft/cluster-api-inplace-updater/pkg/nodes"
)

// test cases
// case: controlplane 0/3 node updated
// case: controlplane 1/3 node updated
// case: controlplane 0/3 node updated, 1 unhealthy
// case: controlplane 0/3 node updated, 1 update fail
// case: controlplane 1/3 node updated, abort
// case: controlplane 1/3 node updated, 1 updating, abort
// case: controlplane 3/3 node updated
// case: machinedeployment 1/3 node updated
// case: not in scope
// case: TBD

const (
	DefaultNamepace = "default"
)

type ExpectedInfo struct {
	StatusState        updatev1beta1.UpdateTaskState
	StatusMessage      *string
	UpdateReason       *string
	AbortReason        *string
	TrackStatusMessage *string
}

type ClusterInfo struct {
	Name               string
	ControlPlane       ControlPlaneInfo
	MachineDeployments []MachineDeploymentInfo
}

type ControlPlaneInfo struct {
	Name     string
	Type     string
	Machines []MachineInfo
}

type MachineDeploymentInfo struct {
	Name     string
	Machines []MachineInfo
}

type MachineInfo struct {
	Name       string
	Conditions []clusterv1.Condition
}

type UpdateTaskInfo struct {
	Name                    string
	UpdateControlPlane      bool
	UpdateMachineDeployment bool
	Phase                   updatev1beta1.UpdateTaskPhase
	Template                *updatev1beta1.FakeNodeUpdateTaskTemplate
	NodeTasks               []NodeUpdateTaskInfo
}

type NodeUpdateTaskInfo struct {
	MachineName string
	Phase       updatev1beta1.UpdateTaskPhase
	State       updatev1beta1.UpdateTaskState
}

func Test1NodeUpdatedNotInScopeReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
		},
	}

	changeScope := func(c client.Client) error {
		machineList := &clusterv1.MachineList{}
		err := c.List(context.TODO(), machineList)
		for _, m := range machineList.Items {
			m.Annotations[updatev1beta1.UpdateTaskAnnotationName] = ""
			if err := c.Update(context.TODO(), &m); err != nil {
				return err
			}
		}
		return err
	}

	extraVerify := func(c client.Client) {
		g := NewWithT(t)
		nodeUpdateList := &updatev1beta1.FakeNodeUpdateTaskList{}
		err := c.List(context.TODO(), nodeUpdateList)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(len(nodeUpdateList.Items)).To((Equal(1)))
	}

	expected := &ExpectedInfo{
		StatusState: updatev1beta1.UpdateTaskStateUpdated,
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, changeScope, extraVerify)
}

func Test3NodeUpdatedReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
			{
				MachineName: "cluster1-cp-2",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
			{
				MachineName: "cluster1-cp-3",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState: updatev1beta1.UpdateTaskStateUpdated,
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test1NodeUpdated1NodeUpdatingAbortReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseAbort,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
			{
				MachineName: "cluster1-cp-2",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateInProgress,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState:   updatev1beta1.UpdateTaskStateInProgress,
		StatusMessage: ptr.To("1 updated, 0 failed, 0 aborted, 1 aborting, 1 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test1NodeUpdatedAbortReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseAbort,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState: updatev1beta1.UpdateTaskStateAborted,
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test1NodeUpdateFailedReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateFailed,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState:        updatev1beta1.UpdateTaskStateInProgress,
		UpdateReason:       ptr.To(updatev1beta1.NodeUpdateFailedReason),
		TrackStatusMessage: ptr.To("0 updated, 1 failed, 0 aborted, 0 inProgress, 2 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func TestUnhealthyControlPlaneReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	unhealthConditions := unhealthyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: unhealthConditions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks:          []NodeUpdateTaskInfo{},
	}

	expected := &ExpectedInfo{
		StatusState:        updatev1beta1.UpdateTaskStateInProgress,
		StatusMessage:      ptr.To("controlplane preflight check failed"),
		UpdateReason:       ptr.To(updatev1beta1.PreflightCheckFailedReason),
		TrackStatusMessage: ptr.To("0 updated, 0 failed, 0 aborted, 0 inProgress, 3 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test0NodeUpdatedReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks:          []NodeUpdateTaskInfo{},
	}

	expected := &ExpectedInfo{
		StatusState:   updatev1beta1.UpdateTaskStateInProgress,
		StatusMessage: ptr.To("0 updated, 0 failed, 0 aborted, 1 inProgress, 2 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test1NodeUpdatedReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:               fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateControlPlane: true,
		Phase:              updatev1beta1.UpdateTaskPhaseUpdate,
		Template:           fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster1-cp-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState:   updatev1beta1.UpdateTaskStateInProgress,
		StatusMessage: ptr.To("1 updated, 0 failed, 0 aborted, 1 inProgress, 1 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func Test1MachineDeploymentNodeUpdatedReconcile(t *testing.T) {
	goodCondtions := readyControlPlaneConditions()
	clusterInfo := &ClusterInfo{
		Name: "cluster1",
		ControlPlane: ControlPlaneInfo{
			Name: "cluster1-cp",
			Type: "KubeadmControlPlane",
			Machines: []MachineInfo{
				{
					Name:       "cluster1-cp-1",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-2",
					Conditions: goodCondtions,
				},
				{
					Name:       "cluster1-cp-3",
					Conditions: goodCondtions,
				},
			},
		},
		MachineDeployments: []MachineDeploymentInfo{
			{
				Name: "cluster-md1",
				Machines: []MachineInfo{
					{
						Name:       "cluster-md1-1",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-2",
						Conditions: goodCondtions,
					},
					{
						Name:       "cluster-md1-3",
						Conditions: goodCondtions,
					},
				},
			},
		},
	}

	taskInfo := &UpdateTaskInfo{
		Name:                    fmt.Sprintf("update-%s-%s", clusterInfo.Name, util.RandomString(6)),
		UpdateMachineDeployment: true,
		Phase:                   updatev1beta1.UpdateTaskPhaseUpdate,
		Template:                fakeNodeUpdateTaskTemplate(),
		NodeTasks: []NodeUpdateTaskInfo{
			{
				MachineName: "cluster-md1-1",
				Phase:       updatev1beta1.UpdateTaskPhaseUpdate,
				State:       updatev1beta1.UpdateTaskStateUpdated,
			},
		},
	}

	expected := &ExpectedInfo{
		StatusState:   updatev1beta1.UpdateTaskStateInProgress,
		StatusMessage: ptr.To("1 updated, 0 failed, 0 aborted, 1 inProgress, 1 waitForUpdate"),
	}

	RunTestCase(t, clusterInfo, taskInfo, expected, nil, nil)
}

func RunTestCase(t *testing.T, clusterInfo *ClusterInfo, taskInfo *UpdateTaskInfo, expected *ExpectedInfo, preSetupFunc func(client.Client) error, extraVerifyFunc func(client.Client)) {
	ctrl.SetLogger(klog.Background())
	g := NewWithT(t)

	ctx := context.TODO()
	client := initialize(*clusterInfo, *taskInfo)

	if preSetupFunc != nil {
		err := preSetupFunc(client)
		g.Expect(err).NotTo(HaveOccurred())
	}

	r := &UpdateTaskReconciler{
		Client: client,
	}

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "default", Name: taskInfo.Name}}
	_, err := reconcile(ctx, r, request, 5)
	g.Expect(err).NotTo(HaveOccurred())

	task := &updatev1beta1.UpdateTask{}
	err = client.Get(ctx, types.NamespacedName{Namespace: DefaultNamepace, Name: taskInfo.Name}, task)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(task).NotTo(BeNil())

	g.Expect(task.Status.State).To(Equal(expected.StatusState))

	if expected.AbortReason != nil {
		g.Expect(conditions.GetReason(task, updatev1beta1.AbortOperationCondition)).To(Equal(*expected.AbortReason))
	}

	if expected.UpdateReason != nil {
		g.Expect(conditions.GetReason(task, updatev1beta1.UpdateOperationCondition)).To(Equal(*expected.UpdateReason))
	}

	if expected.StatusMessage != nil {
		g.Expect(conditions.GetMessage(task, clusterv1.ReadyCondition)).To(Equal(*expected.StatusMessage))
	}

	if expected.TrackStatusMessage != nil {
		g.Expect(conditions.GetMessage(task, updatev1beta1.TrackOperationCondition)).To(Equal(*expected.TrackStatusMessage))
	}

	if extraVerifyFunc != nil {
		extraVerifyFunc(client)
	}
}

func reconcile(ctx context.Context, r *UpdateTaskReconciler, request ctrl.Request, times int) (ctrl.Result, error) {
	var result ctrl.Result
	var err error
	for i := 0; i < times; i++ {
		result, err = r.Reconcile(ctx, request)
	}
	return result, err
}

func unhealthyControlPlaneConditions() []clusterv1.Condition {
	return []clusterv1.Condition{
		{
			Type:   clusterv1.ReadyCondition,
			Status: corev1.ConditionFalse,
		},
		{
			Type:   controlplanev1.MachineAPIServerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineControllerManagerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineSchedulerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:    controlplanev1.MachineEtcdPodHealthyCondition,
			Reason:  "EtcdUnhealthy",
			Message: "Etcd is unhealhty",
			Status:  corev1.ConditionFalse,
		},
		{
			Type:   controlplanev1.MachineEtcdMemberHealthyCondition,
			Status: corev1.ConditionTrue,
		},
	}
}

func readyControlPlaneConditions() []clusterv1.Condition {
	return []clusterv1.Condition{
		{
			Type:   clusterv1.ReadyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineAPIServerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineControllerManagerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineSchedulerPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineEtcdPodHealthyCondition,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   controlplanev1.MachineEtcdMemberHealthyCondition,
			Status: corev1.ConditionTrue,
		},
	}
}

func fakeNodeUpdateTaskTemplate() *updatev1beta1.FakeNodeUpdateTaskTemplate {
	return &updatev1beta1.FakeNodeUpdateTaskTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FakeNodeUpdateTaskTemplate",
			APIVersion: updatev1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "nodeupdatetasktemplate",
			Namespace:   DefaultNamepace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: updatev1beta1.FakeNodeUpdateTaskTemplateSpec{
			Template: updatev1beta1.FakeNodeUpdateTaskTemplateResource{
				Spec: updatev1beta1.FakeNodeUpdateTaskSpec{
					Field1: "f1",
					Field2: "f2",
					Field3: "f3",
					Phase:  updatev1beta1.UpdateTaskPhaseUpdate,
				},
			},
		},
	}
}

func initialize(clusterInfo ClusterInfo, taskInfo UpdateTaskInfo) client.Client {
	f := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	clusterv1.AddToScheme(scheme)
	controlplanev1.AddToScheme(scheme)
	updatev1beta1.AddToScheme(scheme)
	f.WithScheme(scheme)

	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        clusterInfo.Name,
			Namespace:   DefaultNamepace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}
	f.WithObjects(cluster)

	cp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterInfo.ControlPlane.Name,
			Namespace: DefaultNamepace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: cluster.Name,
			},
			Annotations: map[string]string{},
		},
	}
	f.WithObjects(cp)

	for _, mk := range clusterInfo.ControlPlane.Machines {
		machine := &clusterv1.Machine{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Machine",
				APIVersion: clusterv1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      mk.Name,
				Namespace: DefaultNamepace,
				Labels: map[string]string{
					clusterv1.ClusterNameLabel:             cluster.Name,
					clusterv1.MachineControlPlaneLabel:     "",
					clusterv1.MachineControlPlaneNameLabel: cp.Name,
				},
				Annotations: map[string]string{
					updatev1beta1.UpdateTaskAnnotationName: taskInfo.Name,
				},
			},
			Status: clusterv1.MachineStatus{
				NodeRef: &corev1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Node",
					Name:       mk.Name,
				},
				Conditions: mk.Conditions,
			},
		}
		f.WithObjects(machine)
	}

	var firstMd *clusterv1.MachineDeployment = nil
	for _, mdk := range clusterInfo.MachineDeployments {
		md := &clusterv1.MachineDeployment{}
		md.Name = mdk.Name
		md.SetLabels(map[string]string{
			clusterv1.ClusterNameLabel: cluster.Name,
		})
		f.WithObjects(md)
		if firstMd == nil {
			firstMd = md
		}

		for _, mk := range mdk.Machines {
			machine := &clusterv1.Machine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Machine",
					APIVersion: clusterv1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      mk.Name,
					Namespace: DefaultNamepace,
					Labels: map[string]string{
						clusterv1.ClusterNameLabel:           cluster.Name,
						clusterv1.MachineDeploymentNameLabel: md.Name,
					},
					Annotations: map[string]string{
						updatev1beta1.UpdateTaskAnnotationName: taskInfo.Name,
					},
				},
				Status: clusterv1.MachineStatus{
					Conditions: mk.Conditions,
				},
			}
			f.WithObjects(machine)
		}
	}

	updateTask := &updatev1beta1.UpdateTask{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UpdateTask",
			APIVersion: updatev1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      taskInfo.Name,
			Namespace: DefaultNamepace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: cluster.Name,
			},
			Annotations: map[string]string{},
		},
		Spec: updatev1beta1.UpdateTaskSpec{
			ClusterRef:     getReference(cluster),
			TargetPhase:    taskInfo.Phase,
			NewMachineSpec: updatev1beta1.MachineSpec{},
			NodeUpdateTemplate: updatev1beta1.NodeUpdateTaskTemplateSpec{
				InfrastructureRef: *getReference(taskInfo.Template),
			},
		},
	}

	if taskInfo.UpdateControlPlane {
		updateTask.Spec.ControlPlaneRef = getReference(cp)
	}
	if taskInfo.UpdateMachineDeployment {
		updateTask.Spec.MachineDeploymentRef = getReference(firstMd)
	}

	f.WithObjects(updateTask)
	f.WithObjects(taskInfo.Template)

	for _, ni := range taskInfo.NodeTasks {
		spec := taskInfo.Template.Spec.Template.Spec
		spec.Phase = ni.Phase
		spec.MachineName = ni.MachineName
		nodeUpdateTask := &updatev1beta1.FakeNodeUpdateTask{
			TypeMeta: metav1.TypeMeta{
				Kind:       "FakeNodeUpdateTask",
				APIVersion: updatev1beta1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("nodeupdate-%s-%s", ni.MachineName, util.RandomString(6)),
				Namespace: DefaultNamepace,
				Labels: map[string]string{
					clusterv1.ClusterNameLabel:       cluster.Name,
					nodes.ClusterUpdateTaskNameLabel: updateTask.Name,
				},
				Annotations: map[string]string{},
			},
			Spec: spec,
			Status: updatev1beta1.FakeNodeUpdateTaskStatus{
				State: ni.State,
			},
		}
		f.WithObjects(nodeUpdateTask)
	}

	f.WithStatusSubresource(&updatev1beta1.UpdateTask{}, &clusterv1.Machine{})
	return f.Build()
}

func getReference(obj client.Object) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  obj.GetNamespace(),
		Name:       obj.GetName(),
		UID:        obj.GetUID(),
	}
}
