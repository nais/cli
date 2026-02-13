#!/usr/bin/env bash
#MISE description="Run gosec"
set -euo pipefail

go tool github.com/securego/gosec/v2/cmd/gosec --exclude="G204" --exclude-generated -terse ./...
