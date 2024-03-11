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

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"

	upgradev1beta1 "github.com/mogliang/cluster-api-inplace-upgrader/api/v1beta1"
	"github.com/mogliang/cluster-api-inplace-upgrader/internal/controller"
	"github.com/mogliang/cluster-api-inplace-upgrader/internal/handlers"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	runtimecatalog "sigs.k8s.io/cluster-api/exp/runtime/catalog"
	runtimehooksv1 "sigs.k8s.io/cluster-api/exp/runtime/hooks/api/v1alpha1"
	"sigs.k8s.io/cluster-api/exp/runtime/server"
	"sigs.k8s.io/cluster-api/util/flags"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	// catalog contains all information about RuntimeHooks.
	catalog              = runtimecatalog.New()
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string
	secureMetrics        bool
	enableHTTP2          bool
	webhookPort          int
	webhookCertDir       string
	tlsOptions           = flags.TLSOptions{}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(upgradev1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	// Register the RuntimeHook types into the catalog.
	_ = runtimehooksv1.AddToCatalog(catalog)
}

// InitFlags initializes the flags.
func InitFlags(fs *pflag.FlagSet) {
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	flag.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook Server port")
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/",
		"Webhook cert dir, only used when webhook-port is specified.")
	flags.AddTLSOptions(fs, &tlsOptions)
}

func main() {
	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Set log level 2 as default.
	if err := pflag.CommandLine.Set("v", "2"); err != nil {
		setupLog.Error(err, "failed to set log level: %v")
		os.Exit(1)
	}
	pflag.Parse()

	ctrl.SetLogger(klog.Background())

	tlsOptionOverrides, err := flags.GetTLSOptionOverrideFuncs(tlsOptions)
	if err != nil {
		setupLog.Error(err, "unable to add TLS settings to the webhook server")
		os.Exit(1)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOptionOverrides,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOptionOverrides,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5bb3d2b0.github.com/mogliang/",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create an HTTP server for serving Runtime Extensions.
	runtimeExtensionWebhookServer, err := server.New(server.Options{
		Port:    webhookPort,
		CertDir: webhookCertDir,
		TLSOpts: tlsOptionOverrides,
		Catalog: catalog,
	})
	if err != nil {
		setupLog.Error(err, "error creating runtime extension webhook server")
		os.Exit(1)
	}

	setupLifecycleHookHandlers(mgr, runtimeExtensionWebhookServer)

	if err = (&controller.UpgradeTaskReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "UpgradeTask")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// setupLifecycleHookHandlers sets up Lifecycle Hooks.
func setupLifecycleHookHandlers(mgr ctrl.Manager, runtimeExtensionWebhookServer *server.Server) {
	lifecycleExtensionHandlers := handlers.NewExtensionHandlers(mgr.GetClient())

	if err := runtimeExtensionWebhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        upgradev1beta1.ControlPlaneExternalStrategy,
		Name:        "call-controlplane-external-upgrade",
		HandlerFunc: lifecycleExtensionHandlers.HandleControlPlaneExternalStrategy,
	}); err != nil {
		setupLog.Error(err, "error adding handler")
		os.Exit(1)
	}

	if err := runtimeExtensionWebhookServer.AddExtensionHandler(server.ExtensionHandler{
		Hook:        upgradev1beta1.MachineDeploymentExternalStrategy,
		Name:        "call-machinedeployment-external-upgrade",
		HandlerFunc: lifecycleExtensionHandlers.HandleMachineDeploymentExternalStrategy,
	}); err != nil {
		setupLog.Error(err, "error adding handler")
		os.Exit(1)
	}
}
