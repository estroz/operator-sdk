# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
  BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
  Q =
else
  Q = @
endif

export CGO_ENABLED := 0

GIT_VERSION = $(shell git describe --dirty --tags --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
K8S_VERSION = v1.18.2
GOLANGCI_LINT_VER = 1.30.0
OLM_VERSION = 0.15.1
REPO = github.com/operator-framework/operator-sdk
IMAGE_REPO ?= quay.io/operator-framework
PKGS = $(shell go list ./... | grep -v /vendor/)
TEST_PKGS = $(shell go list ./... | grep -v -E 'github.com/operator-framework/operator-sdk/test/')
SOURCES = $(shell find . -name '*.go' -not -path "*/vendor/*")
BUILDPLATFORM ?= $(shell go env GOOS)/$(shell go env GOARCH)
# GO_BUILD_ARGS should be set when running 'go build' or 'go install'.
GO_BUILD_ARGS = \
  -gcflags "all=-trimpath=$(shell go env GOPATH)" \
  -asmflags "all=-trimpath=$(shell go env GOPATH)" \
  -ldflags " \
    -X '$(REPO)/internal/version.GitVersion=$(GIT_VERSION)' \
    -X '$(REPO)/internal/version.GitCommit=$(GIT_COMMIT)' \
  " \

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##############################
# Development                #
##############################

##@ Development

.PHONY: all install

all: format test build/operator-sdk ## Test and Build the Operator SDK

install: ## Install the binaries
	$(Q)$(GOARGS) go install $(GO_BUILD_ARGS) ./cmd/operator-sdk ./cmd/ansible-operator ./cmd/helm-operator

# Code management.
.PHONY: format tidy clean lint setup-k8s

format: ## Format the source code
	$(Q)go fmt $(PKGS)

tidy: ## Update dependencies
	$(Q)go mod tidy -v

clean: ## Clean up the build artifacts
	$(Q)rm -rf build

lint: ## Install and run golangci-lint checks
ifneq ($(GOLANGCI_LINT_VER), "$(shell ./bin/golangci-lint --version 2>/dev/null | cut -b 27-32)")
	@echo "golangci-lint missing or not version '$(GOLANGCI_LINT_VER)', downloading..."
	curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/v$(GOLANGCI_LINT_VER)/install.sh" | sh -s -- -b ./bin "v$(GOLANGCI_LINT_VER)"
endif
	./bin/golangci-lint --timeout 5m run

setup-k8s: ## Set up a kind cluster locally
	./hack/ci/setup-k8s.sh $(K8S_VERSION)

##############################
# Generate Artifacts         #
##############################

##@ Generate

.PHONY: generate cli-doc changelog samples

generate: cli-doc samples bindata ## Run all non-release generate targets

cli-doc: ## Generate CLI documentation
	./hack/generate/cli-doc/cli-doc.sh

samples: ## Generate samples
	go run ./hack/generate/samples/generate_all.go

bindata: ## Generate bindata
	./hack/generate/olm_bindata.sh $(OLM_VERSION)

build/%: $(SOURCES) ## Build the operator-sdk binary
	$(Q)$(GOARGS) go build $(GO_BUILD_ARGS) -o $@ ./cmd/$(patsubst build/,,$*)

# TODO(estroz): inject GOMODCACHE into image builds as a volume to avoid re-pulling modules each build.

# Image build.
image/%:
ifeq ($(findstring scorecard,$*),)
	$(MAKE) build/$*
	docker build -f ./images/$*/Dockerfile -t $(IMAGE_REPO)/$*:dev --build-arg BIN=build/$* .
else
	docker build -f ./images/$*/Dockerfile -t $(IMAGE_REPO)/$*:dev --build-arg BUILDPLATFORM=$(BUILDPLATFORM) .
endif

##############################
# Tests                      #
##############################

##@ Tests

# Static tests.
.PHONY: test test-sanity test-unit

test: test-unit ## Run the tests

test-sanity: tidy build/operator-sdk lint
	./hack/tests/sanity-check.sh

test-unit: ## Run the unit tests
	$(Q)go test -coverprofile=coverage.out -covermode=count -count=1 -short $(TEST_PKGS)

test-links:
	./hack/check-links.sh

# CI tests.
.PHONY: test-ci

test-ci: test-sanity test-unit install test-subcommand test-e2e ## Run the CI test suite

# Subcommand tests.
.PHONY: test-subcommand test-subcommand-olm-install

test-subcommand: test-subcommand-olm-install

test-subcommand-olm-install:
	./hack/tests/subcommand-olm-install.sh

# E2E tests.
.PHONY: test-e2e test-e2e-go test-e2e-ansible test-e2e-ansible-molecule test-e2e-helm

test-e2e: test-e2e-go test-e2e-ansible test-e2e-ansible-molecule test-e2e-helm ## Run the e2e tests

test-e2e-go: image/scorecard-test image/custom-scorecard-tests
	./hack/tests/e2e-go.sh

test-e2e-ansible: image/ansible-operator image/scorecard-test
	./hack/tests/e2e-ansible.sh

test-e2e-ansible-molecule: image/ansible-operator
	./hack/tests/e2e-ansible-molecule.sh

test-e2e-helm: image/helm-operator image/scorecard-test
	./hack/tests/e2e-helm.sh

# Integration tests.
.PHONY: test-integration

test-integration: ## Run integration tests
	./hack/tests/integration.sh
