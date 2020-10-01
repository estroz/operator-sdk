#!/usr/bin/env bash

set -eu

# TODO(estroz): if TAG isn't set, dry-run this rule instead of failing.
: ${TAG?} ${K8S_VERSION?}

TMP_CHANGELOG_PATH=changelog-${TAG}.md

# Generate the changelog first so we can pass it to goreleaser.
go run ./hack/generate/changelog/gen-changelog.go -tag=${TAG} -changelog="$TMP_CHANGELOG_PATH"

if [[ ! -f ./bin/goreleaser ]]; then
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh
fi

export GOPATH="$(go env GOPATH)"
export K8S_VERSION=$K8S_VERSION
export GORELEASER_CURRENT_TAG=$TAG
./bin/goreleaser release \
	--release-notes="$TMP_CHANGELOG_PATH" \
	--parallelism 5 \
	\ --snapshot
	\ --skip-publish
	\ --rm-dist

rm -f ./changelog/fragments/!(00-template.yaml)
rm -f "$TMP_CHANGELOG_PATH"
