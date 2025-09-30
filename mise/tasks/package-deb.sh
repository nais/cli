#!/usr/bin/env bash
#MISE description="Package the Nais CLI binary as a Debian package"
#MISE depends=["build", "completions"]
set -euo pipefail

arch="$GOARCH"
ARCH="$arch" GOARCH="" go tool github.com/goreleaser/nfpm/v2/cmd/nfpm package \
  --packager deb \
  --config .nfpm.yaml \
  --target nais-cli_"$arch".deb