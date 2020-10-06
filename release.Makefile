# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
  BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
  Q =
else
  Q = @
endif

SIMPLE_VERSION=$(shell (test "$(shell git describe)" = "$(shell git describe --abbrev=0)" && echo $(shell git describe)) || echo $(shell git describe --abbrev=0)+git)
GIT_VERSION = $(shell git describe --dirty --tags --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
K8S_VERSION = v1.18.2
GOLANGCI_LINT_VER = "1.30.0"
OLM_VERSION = "0.15.1"
REPO = github.com/operator-framework/operator-sdk
PKGS = $(shell go list ./... | grep -v /vendor/)
TEST_PKGS = $(shell go list ./... | grep -v -E 'github.com/operator-framework/operator-sdk/test/')
SOURCES = $(shell find . -name '*.go' -not -path "*/vendor/*")
# GO_BUILD_ARGS should be set when running 'go build' or 'go install'.
GO_BUILD_ARGS = \
  -gcflags "all=-trimpath=$(shell go env GOPATH)" \
  -asmflags "all=-trimpath=$(shell go env GOPATH)" \
  -ldflags " \
    -X '$(REPO)/internal/version.Version=$(GIT_VERSION)' \
    -X '$(REPO)/internal/version.GitVersion=$(GIT_VERSION)' \
    -X '$(REPO)/internal/version.GitCommit=$(GIT_COMMIT)' \
    -X '$(REPO)/internal/version.KubernetesVersion=$(K8S_VERSION)' \
  " \

export CGO_ENABLED:=0
.DEFAULT_GOAL:=help

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build/operator-sdk build/helm-operator build/ansible-operator
build/%: $(SOURCES) ## Build the operator-sdk binary
	$(Q){ \
	cmdpkg=$$(echo $* | sed -E "s/(operator-sdk|ansible-operator|helm-operator).*/\1/"); \
	$(GOARGS) go build $(GO_BUILD_ARGS) -o $@ ./cmd/$$cmdpkg; \
	}

.PHONY: clean
clean: ## Clean up the build artifacts
	$(Q)rm -rf build

.PHONY: changelog
changelog: ## Generate CHANGELOG.md and migration guide updates
	./hack/generate/changelog/gen-changelog.sh

.PHONY: release
release: ## Release the Operator SDK
	TAG=$(TAG) K8S_VERSION=$(K8S_VERSION) ./release.sh

# TODO(estroz): inject GOMODCACHE into image builds as a volume to avoid re-pulling modules each build.

IMAGE_REPO ?= quay.io/operator-framework

PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le,linux/s390x

image/scorecard-test-kuttl: PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le
image/%: init-docker-buildx
	DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build \
		-f ./images/$*/Dockerfile \
		-t $(IMAGE_REPO)/$*:latest \
		-t $(IMAGE_REPO)/$*:$(GIT_VERSION) \
		--platform $(PLATFORMS) \
		--load \
		.

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	# Ensure qemu is in binfmt_misc
	# NOTE: Please always pin this to a digest for predictability/auditability
	# Last updated: 08/21/2020
	docker run --rm --privileged multiarch/qemu-user-static:latest --reset -p yes

	# Ensure we use a builder that can leverage it (the default on linux will not)
	docker buildx rm operator-sdk-multiarch-builder 2>&1 > /dev/null || true
	docker buildx create --name operator-sdk-multiarch-builder --use
	docker buildx inspect --bootstrap
