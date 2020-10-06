#!/usr/bin/env bash

set -eu

source hack/lib/image_lib.sh

# check_docker_version exits 1 if docker version is less than 19.03, the version
# required to use the buildx plugin and that has it installed by default.
function check_docker_version() {
  local client_version="$(docker version -f {{.Client.Version}})"
  if [[ "$(echo -e "${client_version}\n19.03" | sort -V | tail -n 1)" -eq "19.03" ]]; then
    >&2 echo "docker client version $client_version must be >= 19.03 to use the buildx plugin"
    exit 1
  fi
}

# All parameters.
PARAMS=""
# Comma-separated list of platforms, in the format: "${os}/${arch}".
# If empty, the current arch is used.
PLATFORMS=
# Path to Dockerfile used to build an image.
# If empty, defaults to the current working directory's Dockerfile.
DOCKERFILE=
# Comma-separated list of image tags to build/push.
# This list is required.
TAGS=
# If set, all build images will be pushed (assuming registry access is configured).
PUSH=

set -x

while (( "$#" )); do
  case "$1" in
    -p|--platforms)
      PLATFORMS="$2"
      shift
      ;;
    -f|--dockerfile)
      DOCKERFILE="$2"
      shift
      ;;
    -t|--tags)
      TAGS="$2"
      shift
      ;;
    --push)
      PUSH=true
      ;;
    -*|--*=) # unsupported flags
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
  shift
done
# set positional arguments in their proper place
eval set -- "$PARAMS"

: ${TAGS:?"--tags must be set"}

# Split first tag which will be used to build the image.
declare TAG_LIST
IFS=',' read -r -a TAG_LIST <<< "$TAGS"
FIRST_TAG="${TAG_LIST[0]}"
unset TAG_LIST[0]

# Set the default Dockerfile path for convenience.
[[ -z "$DOCKERFILE" ]] && DOCKERFILE="Dockerfile"

if [[ -z "$PLATFORMS" ]]; then
  # Run a typical build for the current platform if no target platforms are set.
  docker build -f $DOCKERFILE -t $FIRST_TAG .
else
  # Check if buildx is supported, and use it if so.
  check_docker_version
  export DOCKER_CLI_EXPERIMENTAL=enabled
  docker run --privileged --rm tonistiigi/binfmt --install all
  docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
  docker buildx rm operator-sdk-builder 2>&1 > /dev/null || true
  docker buildx create --name operator-sdk-builder --use
  docker buildx build --load -f $DOCKERFILE -t $FIRST_TAG --platform $PLATFORMS .
fi

# Create the rest of the specified tags.
for tag in "${TAG_LIST[@]}"; do
  docker tag $FIRST_TAG $tag
done

# Push images remotely. This assumes DOCKER_USERNAME and DOCKER_PASSWORD have been set externally.
if [[ -n "$PUSH" ]]; then
  docker_login $FIRST_TAG
  docker push $FIRST_TAG
  for tag in "${TAG_LIST[@]}"; do
    docker push $tag
  done
fi
