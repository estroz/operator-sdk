---
title: Release Guide
linkTitle: Releasing
weight: 30
---

These steps describe how to conduct a release of the operator-sdk repo using example versions.
Replace these versions with the current and new version you are releasing, respectively.

## Prerequisites

- [`git`][git]
- [`gpg2`][gpg2] and a [GPG key][gpg-key-create]. If you have both `gpg` and `gpg2` available, make sure the latter is used:
    ```sh
    git config --global gpg.program gpg2
    ```
- Your GPG key is publicly available in a [public key server][gpg-upload], like https://keyserver.ubuntu.com/.

##### MacOS users

Install GNU `sed`, `make`, and `gpg2` which may not be by default:

```sh
brew install gnu-sed make gnupg
```

Configure `gpg2` to sign with `git`:

```bash
echo "export GPG_TTY=tty" >> ~/.bashrc
source ~/.bashrc
```

## Major and Minor releases

We will use the `v1.3.0` release version in this example.

### Before starting

A release branch must be created and [mapped][netlify-deploy] _before the release begins_
to appease the Netlify website configuration demons. You can ping SDK [approvers][doc-owners] to ensure a
[release branch](#release-branches) is created prior to the release and that this mapping is created.
If you have the proper permissions, you can do this by running the following,
assuming the upstream SDK is the `upstream` remote repo:

```sh
git checkout master
git pull
git checkout -b v1.3.x
git push -u upstream v1.3.x
```

### 1. Create and push a release commit

Create a new branch to push the release commit:

```sh
git checkout master
git pull
git checkout -b release-v1.3.0
```

Run the pre-release `make` target:

```sh
make prerelease RELEASE_VERSION=v1.3.0
```

The following changes should be present:

- `changelog/generated/v1.3.0.md`: commit changes (created by changelog generation).
- `changelog/fragments/*`: commit deleted fragment files (deleted by changelog generation).
- `website/content/en/docs/upgrading-sdk-version/v1.3.0.md`: commit changes (created by changelog generation).
- `website/config.toml`: commit changes (modified by release script).

Commit these changes and push:

```sh
git add --all
git commit -m "Release v1.3.0"
git push -u origin release-v1.3.0
```

### 2. Create and merge a new PR

Create and merge a new PR for the above commit.

### 3. Lock down branch `master`

Lock down the `master` branch to prevent further commits before the release completes.
See [this section](#locking-down-branches) for steps to do so.

### 4. Create and push a release tag

```sh
make tag RELEASE_VERSION=v1.3.0
git push upstream v1.3.0
```

### 5. Fast-forward the `latest` and release branches

The `latest` branch points to the latest release tag to keep the main website subdomain up-to-date.
Run the following commands to do so:

```sh
git checkout latest
git reset --hard tags/v1.3.0
git push -f upstream latest
```

Similarly, to update the release branch, run:

```sh
git checkout v1.3.x
git reset --hard tags/v1.3.0
git push -f upstream v1.3.x
```

### 6. Unlock the `master` branch

See [this guide](#unlocking-branches) for steps to do so.

### 7. Post release steps

See the [post-release section](#post-release-steps).


## Patch releases

We will use the `v1.3.1` release version in this example.

### 1. Create and push a release commit

Create a new branch from the release branch, which should already exist for the desired minor version,
to push the release commit to:

```sh
git checkout v1.3.x
git pull
git checkout -b release-v1.3.1
```

Run the pre-release `make` target:

```sh
make prerelease RELEASE_VERSION=v1.3.1
```

The following changes should be present:

- `changelog/generated/v1.3.0.md`: commit changes (created by changelog generation).
- `changelog/fragments/*`: commit deleted fragment files (deleted by changelog generation).

Commit these changes and push:

```sh
git add --all
git commit -m "Release v1.3.1"
git push -u origin release-v1.3.1
```

### 2. Create and merge a new PR

Create and merge a new PR for the above commit.

### 3. Lock down the `v1.3.x` branch

Lock down this branch prevents further commits before the release completes.
See [this section](#locking-down-branches) for steps to do so.

### 4. Create and push a release tag

```sh
make tag RELEASE_VERSION=v1.3.1
git push upstream v1.3.1
```

### 5. Fast-forward the `latest` branch

The `latest` branch points to the latest release tag to keep the main website subdomain up-to-date.
Run the following commands to do so:

```sh
git checkout latest
git reset --hard tags/v1.3.1
git push -f upstream latest
```

### 6. Unlock the `v1.3.x` branch

See [this guide](#unlocking-branches) for steps to do so.

### 7. Post release steps

See the [post-release section](#post-release-steps).

## Further reading

### Binaries and signatures

Binaries will be signed using our CI system's GPG key. Both binary and signature will be uploaded to the release.

### Release branches

Each minor release has a corresponding release branch of the form `vX.Y.x`, where `X` and `Y` are the major and minor
release version numbers and the `x` is literal. This branch accepts bug fixes according to our [backport policy][backports].

##### Cherry-picking

Once a minor release is complete, bug fixes can be merged into the release branch for the next patch release.
Fixes can be added automatically by posting a `/cherry-pick v1.3.x` comment in the `master` PR, or manually by running:

```sh
git checkout v1.3.x
git checkout -b cherrypick/some-bug
git cherry-pick <commit>
git push upstream cherrypick/some-bug
```

Create and merge a PR from your branch to `v1.3.x`.

### GitHub release information

GitHub releases live under the [`Releases` tab][release-page] in the operator-sdk repo.

##### Locking down branches

To lock down a branch:

1. Go to `Settings -> Branches` in the SDK repo.
1. Under `Branch protection rules`, click `Edit` on the `master` or release branches rule.
1. In section `Protect matching branches` of the `Rule settings` box,
increase the number of required approving reviewers to its maximum allowed value.

Now only administrators (maintainers) should be able to force merge PRs.
Make sure everyone in the relevant Slack channel is aware of the release so they do not force merge by accident.

##### Unlocking branches

Unlock a branch by changing the number of required approving reviewers back to 1.

### Post-release steps

##### Announce the release

Send an email to the [mailing list][mailing-list].
Post to Kubernetes slack in #kubernetes-operators and #operator-sdk-dev.

##### Bump open issues to the next release.

In the [GitHub milestone][gh-milestones], bump any open issues to the
following release.


[git]:https://git-scm.com/downloads
[gpg2]:https://gnupg.org/download/
[gpg-key-create]:https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/managing-commit-signature-verification
[gpg-upload]:https://www.gnupg.org/gph/en/manual/x457.html
[netlify-deploy]:https://docs.netlify.com/site-deploys/overview/#deploy-summary
[doc-owners]: https://github.com/operator-framework/operator-sdk/blob/master/OWNERS
[release-page]:https://github.com/operator-framework/operator-sdk/releases
[backports]:/docs/upgrading-sdk-version/backport-policy
[mailing-list]:https://groups.google.com/g/operator-framework
[gh-milestones]:https://github.com/operator-framework/operator-sdk/milestones
