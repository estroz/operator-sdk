K8S_VERSION = v1.18.2

# Image variables.
IMAGE_REPO ?= quay.io/operator-framework
PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le,linux/s390x
# kuttl doesn't support s390x yet.
KUTTL_PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: check_tag
check_tag:
ifeq ($(TAG),)
	$(error "TAG must be set to a release tag")
endif

.PHONY: snapshot
snapshot: goreleaser ## Release the Operator SDK
	export K8S_VERSION=$(K8S_VERSION)
	./bin/goreleaser release --snapshot --skip-publish --rm-dist --parallelism 5

.PHONY: release
release: check_tag goreleaser ## Release the Operator SDK
	go run ./release/changelog/gen-changelog.go -tag=$(TAG) -changelog=changelog-$(TAG).md
	export GORELEASER_CURRENT_TAG=$(TAG)
	export K8S_VERSION=$(K8S_VERSION)
	./bin/goreleaser release --release-notes=changelog-$(TAG).md --parallelism 5
	bash -c "rm -f ./changelog/fragments/!(00-template.yaml)"
	rm -f changelog-$(TAG).md

# TODO(estroz): inject GOMODCACHE into image builds as a volume to avoid re-pulling modules each build.

.PHONY: goreleaser
goreleaser:
	[[ ! -f ./bin/goreleaser ]] && curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh || true

image/%: check_tag
	$(MAKE) -f ./images/Makefile $@ IMAGE_TAGS="$(IMAGE_REPO)/$*:latest $(IMAGE_REPO)/$*:$(TAG)"
