#!/usr/bin/env bash
function validate_title {
	grep -qE '^(Merge|((feat|fix|ci|docs|refactor|perf|test|build|style)(\([a-z0-9\s\-\_\,]+\))?!?:\s\w))' <<<"$1"
}

function explain_scheme {
	echo "Your commit message did not follow semantic versioning: $1"
	echo ""
	echo "Format:   <type>(<optional-scope>): <subject>"
	echo "Example:  feat(api): add endpoint"
	echo ""
	echo "Type     | Description"
	echo "---------+------------"
	echo "feat     | Introduces a new feature"
	echo "fix      | Patches a bug"
	echo "ci       | CI configuration files and scripts (i.e. .github/**, some mise tasks)"
	echo "docs     | Documentation only changes (i.e. README, code comments)"
	echo "refactor | Neither bugfix nor adds a feature (i.e. rename package, move code"
	echo "perf     | Improves performance (i.e. removes a time.Sleep)"
	echo "test     | Adding / correcting tests"
	echo "build    | Build system or external dependencies (i.e. go.mod, mise tasks)"
	echo "style    | Changes to output formatting / colors (i.e. changing wording in an error)"
	echo ""
	echo "Please see"
	echo "- https://www.conventionalcommits.org/en/v1.0.0/#summary"
}
