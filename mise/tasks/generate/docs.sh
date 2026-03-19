#!/usr/bin/env bash
#MISE description="Generate Markdown docs (cli.nais.io)"
set -euo pipefail

rm -f "$MISE_PROJECT_ROOT"/docs/nais_*.md || true
go run cmd/gen_docs/main.go
