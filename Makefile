
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Cluster API: Makefile

#
# Go.
#
GO_VERSION ?= 1.21.8
GO_DIRECTIVE_VERSION ?= 1.21
GO_CONTAINER_IMAGE ?= docker.io/library/golang:$(GO_VERSION)

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

#
# Kubebuilder.
#
export KUBEBUILDER_ENVTEST_KUBERNETES_VERSION ?= 1.29.0
export KUBEBUILDER_CONTROLPLANE_START_TIMEOUT ?= 60s
export KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT ?= 60s

# This option is for running docker manifest command
export DOCKER_CLI_EXPERIMENTAL := enabled

# Enables shell script tracing. Enable by running: TRACE=1 make <target>
TRACE ?= 0

#
# Directories.
#
# Full directory of where the Makefile resides
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
CAPI_ROOT_DIR:=$(ROOT_DIR)/cluster-api
EXP_DIR := exp
BIN_DIR := bin
TEST_DIR := test
TOOLS_DIR := hack/tools
CAPI_TOOLS_BIN_DIR := $(abspath cluster-api/$(TOOLS_DIR)/$(BIN_DIR))
DOCS_DIR := docs
E2E_FRAMEWORK_DIR := $(TEST_DIR)/framework
CAPD_DIR := $(TEST_DIR)/infrastructure/docker
CAPIM_DIR := $(TEST_DIR)/infrastructure/inmemory
TEST_EXTENSION_DIR := cluster-api/$(TEST_DIR)/extension
GO_INSTALL := ./cluster-api/scripts/go_install.sh
OBSERVABILITY_DIR := cluster-api/hack/observability

export PATH := $(abspath $(CAPI_TOOLS_BIN_DIR)):$(PATH)

# Set --output-base for conversion-gen if we are not within GOPATH
ifneq ($(abspath $(CAPI_ROOT_DIR)),$(shell go env GOPATH)/src/sigs.k8s.io/cluster-api)
	CONVERSION_GEN_OUTPUT_BASE := --output-base=$(CAPI_ROOT_DIR)
	CONVERSION_GEN_OUTPUT_BASE_CAPD := --output-base=$(CAPI_ROOT_DIR)/$(CAPD_DIR)
else
	export GOPATH := $(shell go env GOPATH)
endif

#
# Ginkgo configuration.
#
GINKGO_FOCUS ?=
GINKGO_SKIP ?=
GINKGO_NODES ?= 1
GINKGO_TIMEOUT ?= 2h
GINKGO_POLL_PROGRESS_AFTER ?= 60m
GINKGO_POLL_PROGRESS_INTERVAL ?= 5m
E2E_CONF_FILE ?= $(CAPI_ROOT_DIR)/$(TEST_DIR)/e2e/config/docker.yaml
SKIP_RESOURCE_CLEANUP ?= false
USE_EXISTING_CLUSTER ?= false
GINKGO_NOCOLOR ?= false

# to set multiple ginkgo skip flags, if any
ifneq ($(strip $(GINKGO_SKIP)),)
_SKIP_ARGS := $(foreach arg,$(strip $(GINKGO_SKIP)),-skip="$(arg)")
endif

# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')

#
# Binaries.
#
# Note: Need to use abspath so we can invoke these from subdirectories
KUSTOMIZE_VER := v4.5.2
KUSTOMIZE_BIN := kustomize
KUSTOMIZE := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER))
KUSTOMIZE_PKG := sigs.k8s.io/kustomize/kustomize/v4

SETUP_ENVTEST_VER := v0.0.0-20240215143116-d0396a3d6f9f
SETUP_ENVTEST_BIN := setup-envtest
SETUP_ENVTEST := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(SETUP_ENVTEST_BIN)-$(SETUP_ENVTEST_VER))
SETUP_ENVTEST_PKG := sigs.k8s.io/controller-runtime/tools/setup-envtest

CONTROLLER_GEN_VER := v0.14.0
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER))
CONTROLLER_GEN_PKG := sigs.k8s.io/controller-tools/cmd/controller-gen

GOTESTSUM_VER := v1.11.0
GOTESTSUM_BIN := gotestsum
GOTESTSUM := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(GOTESTSUM_BIN)-$(GOTESTSUM_VER))
GOTESTSUM_PKG := gotest.tools/gotestsum

CONVERSION_GEN_VER := v0.29.2
CONVERSION_GEN_BIN := conversion-gen
# We are intentionally using the binary without version suffix, to avoid the version
# in generated files.
CONVERSION_GEN := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(CONVERSION_GEN_BIN))
CONVERSION_GEN_PKG := k8s.io/code-generator/cmd/conversion-gen

ENVSUBST_BIN := envsubst
ENVSUBST_VER := $(call get_go_version,github.com/drone/envsubst/v2)
ENVSUBST := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(ENVSUBST_BIN)-$(ENVSUBST_VER))
ENVSUBST_PKG := github.com/drone/envsubst/v2/cmd/envsubst

GO_APIDIFF_VER := v0.8.2
GO_APIDIFF_BIN := go-apidiff
GO_APIDIFF := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(GO_APIDIFF_BIN)-$(GO_APIDIFF_VER))
GO_APIDIFF_PKG := github.com/joelanford/go-apidiff

HADOLINT_VER := v2.12.0
HADOLINT_FAILURE_THRESHOLD = warning

SHELLCHECK_VER := v0.9.0

TRIVY_VER := 0.49.1

KPROMO_VER := v4.0.5
KPROMO_BIN := kpromo
KPROMO :=  $(abspath $(CAPI_TOOLS_BIN_DIR)/$(KPROMO_BIN)-$(KPROMO_VER))
# KPROMO_PKG may have to be changed if KPROMO_VER increases its major version.
KPROMO_PKG := sigs.k8s.io/promo-tools/v4/cmd/kpromo

YQ_VER := v4.35.2
YQ_BIN := yq
YQ :=  $(abspath $(CAPI_TOOLS_BIN_DIR)/$(YQ_BIN)-$(YQ_VER))
YQ_PKG := github.com/mikefarah/yq/v4

PLANTUML_VER := 1.2024.3

GINKGO_BIN := ginkgo
GINKGO_VER := $(call get_go_version,github.com/onsi/ginkgo/v2)
GINKGO := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINKGO_VER))
GINKGO_PKG := github.com/onsi/ginkgo/v2/ginkgo

GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT_VER := $(shell cat cluster-api/.github/workflows/pr-golangci-lint.yaml | grep [[:space:]]version: | sed 's/.*version: //')
GOLANGCI_LINT := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER))
GOLANGCI_LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

GOVULNCHECK_BIN := govulncheck
GOVULNCHECK_VER := v1.0.4
GOVULNCHECK := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(GOVULNCHECK_BIN)-$(GOVULNCHECK_VER))
GOVULNCHECK_PKG := golang.org/x/vuln/cmd/govulncheck

IMPORT_BOSS_BIN := import-boss
IMPORT_BOSS_VER := v0.28.1
IMPORT_BOSS := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(IMPORT_BOSS_BIN))
IMPORT_BOSS_PKG := k8s.io/code-generator/cmd/import-boss

CONVERSION_VERIFIER_BIN := conversion-verifier
CONVERSION_VERIFIER := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(CONVERSION_VERIFIER_BIN))

OPENAPI_GEN_VER := 70dd376
OPENAPI_GEN_BIN := openapi-gen
# We are intentionally using the binary without version suffix, to avoid the version
# in generated files.
OPENAPI_GEN := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(OPENAPI_GEN_BIN))
OPENAPI_GEN_PKG := k8s.io/kube-openapi/cmd/openapi-gen

PROWJOB_GEN_BIN := prowjob-gen
PROWJOB_GEN := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(PROWJOB_GEN_BIN))

RUNTIME_OPENAPI_GEN_BIN := runtime-openapi-gen
RUNTIME_OPENAPI_GEN := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(RUNTIME_OPENAPI_GEN_BIN))

TILT_PREPARE_BIN := tilt-prepare
TILT_PREPARE := $(abspath $(CAPI_TOOLS_BIN_DIR)/$(TILT_PREPARE_BIN))

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
PROD_REGISTRY ?= registry.k8s.io/cluster-api

STAGING_REGISTRY ?= gcr.io/k8s-staging-cluster-api
STAGING_BUCKET ?= artifacts.k8s-staging-cluster-api.appspot.com

# core
IMAGE_NAME ?= cluster-api-controller
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)

# bootstrap
KUBEADM_BOOTSTRAP_IMAGE_NAME ?= kubeadm-bootstrap-controller
KUBEADM_BOOTSTRAP_CONTROLLER_IMG ?= $(REGISTRY)/$(KUBEADM_BOOTSTRAP_IMAGE_NAME)

# control plane
KUBEADM_CONTROL_PLANE_IMAGE_NAME ?= kubeadm-control-plane-controller
KUBEADM_CONTROL_PLANE_CONTROLLER_IMG ?= $(REGISTRY)/$(KUBEADM_CONTROL_PLANE_IMAGE_NAME)

# capd
CAPD_IMAGE_NAME ?= capd-manager
CAPD_CONTROLLER_IMG ?= $(REGISTRY)/$(CAPD_IMAGE_NAME)

# capim
CAPIM_IMAGE_NAME ?= capim-manager
CAPIM_CONTROLLER_IMG ?= $(REGISTRY)/$(CAPIM_IMAGE_NAME)

# clusterctl
CLUSTERCTL_MANIFEST_DIR := cmd/clusterctl/config
CLUSTERCTL_IMAGE_NAME ?= clusterctl
CLUSTERCTL_IMG ?= $(REGISTRY)/$(CLUSTERCTL_IMAGE_NAME)

# test extension
TEST_EXTENSION_IMAGE_NAME ?= test-extension
TEST_EXTENSION_IMG ?= $(REGISTRY)/$(TEST_EXTENSION_IMAGE_NAME)

# kind
CAPI_KIND_CLUSTER_NAME ?= capi-test

# It is set by Prow GIT_TAG, a git-based tag of the form vYYYYMMDD-hash, e.g., v20210120-v0.3.10-308-gc61521971

TAG ?= dev
ARCH ?= $(shell go env GOARCH)
ALL_ARCH ?= amd64 arm arm64 ppc64le s390x

# Allow overriding the imagePullPolicy
PULL_POLICY ?= Always

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif

##@ Cluster API: Makefile: Testing

ARTIFACTS ?= ${ROOT_DIR}/_artifacts
CAPI_ARTIFACTS ?= ${CAPI_ROOT_DIR}/_artifacts

##@ .PHONY

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: $(CONTROLLER_GEN_BIN) ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="{./api/..., ./cmd/..., ./internal/..., ./stub/...}" output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: $(CONTROLLER_GEN_BIN) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="{./api/..., ./cmd/..., ./internal/..., ./stub/...}"

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

# Utilize Kind or modify the e2e tests to load the image locally, enabling compatibility with other vendors.
.PHONY: test-e2e  # Run the e2e tests against a Kind k8s instance that is spun up.
test-e2e:
	go test ./test/e2e/ -v -ginkgo.v

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN) ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT_BIN) ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	@if [ -d "config/crd" ]; then \
		$(KUSTOMIZE) build config/crd > dist/install.yaml; \
	fi
	echo "---" >> dist/install.yaml  # Add a document separator before appending
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default >> dist/install.yaml

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Cluster API

.PHONY: capi-docker-build-e2e
capi-docker-build-e2e:
	$(MAKE) -C cluster-api docker-build-e2e

.PHONY: capi-test-e2e
capi-test-e2e:
	$(MAKE) -C cluster-api GINKGO_FOCUS="$(GINKGO_FOCUS)" test-e2e

.PHONY: generate-e2e-templates
generate-e2e-templates: capi-generate-e2e-templates

.PHONY: capi-generate-e2e-templates
capi-generate-e2e-templates:
	$(MAKE) -C cluster-api generate-e2e-templates

.PHONY: new-test-e2e
new-test-e2e: $(GINKGO_BIN) generate-e2e-templates ## Run the end-to-end tests
	CNI="../../cluster-api/test/e2e/data/cni/kindnet/kindnet.yaml" \
	KUBETEST_CONFIGURATION="../../cluster-api/test/e2e/data/kubetest/conformance.yaml" \
	AUTOSCALER_WORKLOAD="../../cluster-api/test/e2e/data/autoscaler/autoscaler-to-workload-workload.yaml" \
	$(GINKGO) -v --trace -poll-progress-after=$(GINKGO_POLL_PROGRESS_AFTER) \
		-poll-progress-interval=$(GINKGO_POLL_PROGRESS_INTERVAL) --tags=e2e --focus="$(GINKGO_FOCUS)" \
		$(_SKIP_ARGS) --nodes=$(GINKGO_NODES) --timeout=$(GINKGO_TIMEOUT) --no-color=$(GINKGO_NOCOLOR) \
		--output-dir="$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" $(GINKGO_ARGS) $(ROOT_DIR)/$(TEST_DIR)/e2e -- \
	    -e2e.artifacts-folder="$(CAPI_ARTIFACTS)" \
	    -e2e.config="$(E2E_CONF_FILE)" \
	    -e2e.skip-resource-cleanup=$(SKIP_RESOURCE_CLEANUP) -e2e.use-existing-cluster=$(USE_EXISTING_CLUSTER)

.PHONY: $(CONTROLLER_GEN_BIN)
$(CONTROLLER_GEN_BIN): capi-$(CONTROLLER_GEN_BIN)

.PHONY: capi-$(CONTROLLER_GEN_BIN)
capi-$(CONTROLLER_GEN_BIN):
	$(MAKE) -C cluster-api $(CONTROLLER_GEN_BIN)

.PHONY: $(KUSTOMIZE_BIN)
$(KUSTOMIZE_BIN): capi-$(KUSTOMIZE_BIN)

.PHONY: capi-$(KUSTOMIZE_BIN)
capi-$(KUSTOMIZE_BIN):
	$(MAKE) -C cluster-api $(KUSTOMIZE_BIN)

.PHONY: $(SETUP_ENVTEST_BIN)
$(SETUP_ENVTEST_BIN): capi-$(SETUP_ENVTEST_BIN)

.PHONY: capi-$(SETUP_ENVTEST_BIN)
capi-$(SETUP_ENVTEST_BIN):
	$(MAKE) -C cluster-api $(SETUP_ENVTEST_BIN)

.PHONY: $(GINKGO_BIN)
$(GINKGO_BIN): capi-$(GINKGO_BIN)

.PHONY: capi-$(GINKGO_BIN)
capi-$(GINKGO_BIN):
	$(MAKE) -C cluster-api $(GINKGO_BIN)

.PHONY: $(GOLANGCI_LINT_BIN)
$(GOLANGCI_LINT_BIN): capi-$(GOLANGCI_LINT_BIN)

.PHONY: capi-$(GOLANGCI_LINT_BIN)
capi-$(GOLANGCI_LINT_BIN):
	$(MAKE) -C cluster-api $(GOLANGCI_LINT_BIN)
