#!/usr/bin/env bash
#MISE description="Package the Nais CLI binary as a Debian package"
#MISE depends=["build"]
set -euo pipefail

version="${VERSION:-local}"
arch="$GOARCH"

ARCH="$arch" GOARCH="" go tool github.com/goreleaser/nfpm/v2/cmd/nfpm package \
  --packager deb \
  --config .nfpm.yaml \
  --target nais-cli_"$version"_"$arch".deb