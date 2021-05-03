#! /usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function prerelease() {
    local TARGET_BRANCH=${1:-master}

    git diff --exit-code || (echo "Git state is dirty. Please commit or stash your changes before running prerelease." && exit 1)
    git checkout ${UPSTREAM_REMOTE} ${TARGET_BRANCH}
    git branch release-${RELEASE_VERSION}
    git checkout release-${RELEASE_VERSION}

    sed -i -E 's/(IMAGE_VERSION = ).+/\1'"${RELEASE_VERSION}"'/g' /operator-sdk/Makefile
    make generate prerelease
    git add --all

    echo "The pre-release commit for release ${RELEASE_VERSION} has been prepared."
    echo "Ensure the generated changelog and migration guides have poper formatting, then run:"
    echo "\$ git commit --signoff --message \"Release $RELEASE_VERSION\""
    echo "\$ git push --set-upstream <your-remote> release-${RELEASE_VERSION}"
    echo "where <your-remote> is your personal fork's remote mapping, NOT ${UPSTREAM_REMOTE}."
}

function tag() {
    local TARGET_BRANCH=${1:-master}

    git diff --exit-code || (echo "Git state is dirty. Please commit or stash your changes before tagging." && exit 1)
    git checkout ${UPSTREAM_REMOTE} ${TARGET_BRANCH}
    git pull

    make tag

    echo "Release tag ${RELEASE_VERSION} has been created."
    echo "Run the following to push this tag to the operator-framework/operator-sdk repo:"
    echo "\$ git push ${UPSTREAM_REMOTE} refs/tags/${RELEASE_VERSION}"
    echo "Make sure the workflow run passes before updating latest and patch branches"
}

STEP=$1
export RELEASE_VERSION=${2:?RELEASE_VERSION must be set}
shift 2
export UPSTREAM_REMOTE=$(git remote -v | grep operator-framework/operator-sdk | head -n 1 | awk '{print $1}')
case $STEP in
prerelease) prerelease ;;
tag) tag ;;
*)
echo "$STEP is an invalid step, must be one of: prerelease, tag"
exit 1
;;
esac

