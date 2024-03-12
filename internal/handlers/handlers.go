package handlers

import (
	"context"

	upgradev1beta1 "github.com/mogliang/cluster-api-inplace-upgrader/api/v1beta1"
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

func (m *ExtensionHandlers) HandleControlPlaneExternalStrategy(ctx context.Context, request *upgradev1beta1.ControlPlaneExternalStrategyRequest, response *upgradev1beta1.ExternalStrategyResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("ControlPlaneExternalStrategyRequest is called")

	response.Accepted = false
}

func (m *ExtensionHandlers) HandleMachineDeploymentExternalStrategy(ctx context.Context, request *upgradev1beta1.MachineDeploymentExternalStrategyRequest, response *upgradev1beta1.ExternalStrategyResponse) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("MachineDeploymentExternalStrategyRequest is called")

	response.Accepted = false
}
