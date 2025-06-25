#!/usr/bin/env bash
# Add to git hooks as a pre-commit hook to enforce semantic commit messages.
# ln -s ../../script/semantic-commit-hook.sh .git/hooks/commit-msg

if [ -z "$1" ]; then
	echo "Missing argument (commit message). Did you try to run this manually?"
	exit 1
fi

source "$(dirname "$(readlink -f "$0")")/semantic-commit-scheme.sh"

commit_title="$(head "$1" -n1)"
if ! validate_title "$commit_title"; then
	explain_scheme "$commit_title"
	exit 1
fi
