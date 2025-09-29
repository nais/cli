#!/usr/bin/env bash
#MISE description="Build the Nais CLI binary"
set -euo pipefail

version="${VERSION:-local}"

go build \
  -ldflags "-s -w -X github.com/nais/cli/internal/version.Version=$version" \
  -o bin/nais ./