#!/usr/bin/env bash
#MISE description="Run gosec"
set -euo pipefail

# Most of these exclusions are due to the fact that this is a tool used locally, and not served on the internet. If users of this tool want to hack themselves they are free to do so.
excluded_checks=(
	"G204" # allow passing flags directly to subprocesses
	"G204" # allow subprocesses that use variables
	"G107" # allow user input in http requests
	"G304" # allow path traversal
	"G112" # allow http serve without timeouts - slowloris is not a concern for apps running locally
	"G114" ## allow http serve without timeouts - slowloris is not a concern for apps running locally
)
(
	IFS=','
	go tool github.com/securego/gosec/v2/cmd/gosec \
		-exclude-generated \
		-terse \
		"-exclude=${excluded_checks[*]}" \
		./...
)
