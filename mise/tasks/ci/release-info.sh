#!/usr/bin/env bash
#MISE description="Ensure all code is formatted"
#MISE depends=["fmt"]
set -euo pipefail

repository="${1:-$GITHUB_REPOSITORY}"
token="${2:-$GITHUB_TOKEN}"

version="$(git-cliff --bumped-version)"
echo "version=$version" >>"$GITHUB_OUTPUT"
echo "Bumped version: $version"
changelog="$(git-cliff \
  --tag "$version" \
  --github-repo "$repository" \
  --github-token "$token" \
  --unreleased \
  --strip all \
  -v)"
echo "changelog<<EOF" >>"$GITHUB_OUTPUT"
echo "$changelog" >>"$GITHUB_OUTPUT"
echo "EOF" >>"$GITHUB_OUTPUT"

if [[ "$GITHUB_EVENT_NAME" == "pull_request" ]]; then
  echo -n "PR comment with release info: "
  if [[ -n "$changelog" ]]; then
    pr_comment="# :pencil: Changelog preview
Below is a preview of the Changelog that will be added to the next release. \
Only commit messages that follow the [Conventional Commits specification](https://www.conventionalcommits.org/) will be included in the Changelog.

$changelog"
  else
    pr_comment="# :disappointed: No release for you
There are no commits in your branch that follow the [Conventional Commits specification](https://www.conventionalcommits.org/), so no release will be created.

If you want to create a release from this pull request, please reword your commit messages to replace this message with a preview of a beautiful Changelog."
  fi

  echo -e "$pr_comment" | gh pr comment "${GITHUB_REF_NAME%%/merge}" \
    --edit-last --create-if-none \
    --repo "$repository" \
    --body-file=-
fi