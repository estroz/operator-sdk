# Image variables.
PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le,linux/s390x
# kuttl doesn't support s390x yet.
KUTTL_PLATFORMS := linux/amd64,linux/arm64,linux/ppc64le

# Required for the docker buildx plugin.
export DOCKER_CLI_EXPERIMENTAL := enabled

# Override PLATFORMS for kuttl.
%/scorecard-test-kuttl: PLATFORMS = $(KUTTL_PLATFORMS)
# Default image tag has no registery/repo prefix and a "dev" suffix, for local use only.
image/% image-multiarch/%: IMAGE_TAGS ?= $*:dev
image/%: ## Build an image for the current architecture.
	docker buildx build --progress plain -f ./images/$*/Dockerfile -t $(shell echo "$(IMAGE_TAGS)" | sed 's/ / -t /g') --load .
image-multiarch/%: init-docker-buildx ## Build a manifest list of images for all PLATFORMS.
	docker buildx build --progress plain -f ./images/$*/Dockerfile -t $(shell echo "$(IMAGE_TAGS)" | sed 's/ / -t /g') --push --platform $(PLATFORMS) .

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell DOCKER_CLI_EXPERIMENTAL=enabled docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	# Ensure qemu is in binfmt_misc
	docker run --rm --privileged multiarch/qemu-user-static:latest --reset -p yes
	# Ensure we use a builder that can leverage it (the default on linux will not)
	docker buildx rm operator-sdk-multiarch-builder 2>&1 > /dev/null || true
	docker buildx create --name operator-sdk-multiarch-builder --use --driver-opt image=moby/buildkit:master
	docker buildx inspect --bootstrap
