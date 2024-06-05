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

package v1beta1

type UpdateTaskPhase string

const (
	UpdateTaskAnnotationName = "update.extension.cluster.x-k8s.io/cluster-update-task-name"
)

const (
	UpdateTaskPhaseUpdate UpdateTaskPhase = "Update"
	UpdateTaskPhaseAbort  UpdateTaskPhase = "Abort"
)

type UpdateTaskState string

const (
	UpdateTaskStateUnknown    UpdateTaskState = ""
	UpdateTaskStateInProgress UpdateTaskState = "InProgress"
	UpdateTaskStateFailed     UpdateTaskState = "Failed"
	UpdateTaskStateUpdated    UpdateTaskState = "Updated"
	UpdateTaskStateAborted    UpdateTaskState = "Aborted"
)

func IsTerminated(state UpdateTaskState) bool {
	return state == UpdateTaskStateFailed ||
		state == UpdateTaskStateUpdated ||
		state == UpdateTaskStateAborted
}
