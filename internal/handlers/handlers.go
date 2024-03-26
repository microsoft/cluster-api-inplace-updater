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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/microsoft/cluster-api-inplace-upgrader/stub"
)

type ExtensionHandlers struct {
	client client.Client
}

func NewExtensionHandlers(client client.Client) *ExtensionHandlers {
	return &ExtensionHandlers{
		client: client,
	}
}

func (m *ExtensionHandlers) DoControlPlaneExternalUpgrade(ctx context.Context, request *stub.ControlPlaneExternalUpgradeRequest, response *stub.ControlPlaneExternalUpgradeResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("ControlPlaneExternalUpgrade is called")
}

func (m *ExtensionHandlers) DoMachineDeploymentExternalUpgrade(ctx context.Context, request *stub.MachineDeploymentExternalUpgradeRequest, response *stub.MachineDeploymentExternalUpgradeResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("DoMachineDeploymentExternalUpgrade is called")
}
