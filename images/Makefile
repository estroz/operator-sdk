# Image variables.
PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le,linux/s390x
# kuttl doesn't support s390x yet.
KUTTL_PLATFORMS = linux/amd64,linux/arm64,linux/ppc64le
IMAGE_REPO ?= quay.io/operator-framework
IMAGE_TAG_SUFFIX ?= latest

export DOCKER_CLI_EXPERIMENTAL=enabled

image/scorecard-test-kuttl: PLATFORMS ?= KUTTL_PLATFORMS
image/%: IMAGE_TAGS ?= $(IMAGE_REPO)/$*:$(IMAGE_TAG_SUFFIX)
# image/%: init-docker-buildx
image/%:
	docker buildx build -f ./$*/Dockerfile -t $(shell sed 's/ / -t /g' <<< "$(IMAGE_TAGS)") --push --platform $(PLATFORMS) .

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	# Ensure qemu is in binfmt_misc
	docker run --rm --privileged multiarch/qemu-user-static:latest --reset -p yes
	# Ensure we use a builder that can leverage it (the default on linux will not)
	docker buildx rm operator-sdk-multiarch-builder 2>&1 > /dev/null || true
	docker buildx create --name operator-sdk-multiarch-builder --use
	docker buildx inspect --bootstrap
