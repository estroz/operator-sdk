SHELL = /bin/bash
.SHELLFLAGS = -O extglob -c

# Version variables.
K8S_VERSION = v1.18.2
GIT_VERSION = $(shell git describe --dirty --always --tags)

# Image variables.
IMAGE_REPO ?= quay.io/operator-framework
# Suffixes (the string following ":") of image tags that will be appended to $IMAGE_REPO and image name.
IMAGE_TAG_SUFFIXES ?= $(GIT_VERSION)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: snapshot
snapshot: goreleaser ## Build all operator-sdk project binaries.
	GORELEASER_CURRENT_TAG=$(GIT_VERSION) K8S_VERSION=$(K8S_VERSION) ./bin/goreleaser release --snapshot --skip-publish --rm-dist --parallelism 5

.PHONY: release
release: goreleaser ## Release all operator-sdk binaries with notes. $TAG must be explicitly set to avoid ambiguity.
ifeq ($(TAG),)
	$(error "TAG must be set to a release tag")
endif
	go run ./release/changelog/gen-changelog.go -tag=$(TAG) -changelog=changelog-$(TAG).md
	GORELEASER_CURRENT_TAG=$(TAG) K8S_VERSION=$(K8S_VERSION) ./bin/goreleaser release --release-notes=changelog-$(TAG).md --parallelism 5
	rm -f ./changelog/fragments/!(00-template.yaml)
	rm -f changelog-$(TAG).md

.PHONY: goreleaser
goreleaser: ## Install goreleaser locally.
	@[ ! -f ./bin/goreleaser ] && curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh || true

image/%: ## Build an image with release tags.
	$(MAKE) -f ./images/Makefile image-multiarch/$* IMAGE_TAGS="$(patsubst %,$(IMAGE_REPO)/$*:%,$(IMAGE_TAG_SUFFIXES))"
