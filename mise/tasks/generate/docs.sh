#!/usr/bin/env bash
#MISE description="Generate Markdown docs (cli.nais.io)"
set -euo pipefail

mkdir -p "$MISE_PROJECT_ROOT/docs/_commands"
rm -f "$MISE_PROJECT_ROOT"/docs/nais_*.md "$MISE_PROJECT_ROOT"/docs/_commands/nais_*.md "$MISE_PROJECT_ROOT"/docs/configuration.md || true
go run cmd/gen_docs/main.go
mv "$MISE_PROJECT_ROOT"/docs/nais_*.md "$MISE_PROJECT_ROOT/docs/_commands/"
