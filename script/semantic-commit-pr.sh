#!/usr/bin/env bash

source "$(dirname "$(readlink -f "$0")")/semantic-commit-scheme.sh"
base="${1:-main}"

echo "verifying commit messages against semantic commit scheme..."
git log --format="%s" "$(git merge-base "$base" HEAD)..HEAD"

bad_commits=""
while read -r line; do
	echo "Checking commit: $line"
	if ! validate_title "$line"; then
		bad_commits+="- $line\n"
	fi
done < <(git log --format="%s" "$(git merge-base "$base" HEAD)..HEAD")

if [ -n "$bad_commits" ]; then
	body="### :exclamation: Commits detected that don't follow the commit scheme:\n$bad_commits"
	echo -e "Posting to GitHub PR:\n$body"
	echo -e "$body" | gh pr comment "$GITHUB_PR_NUMBER" --edit-last --create-if-none --body-file=- --repo "$GITHUB_REPO" --
fi

exit 0
