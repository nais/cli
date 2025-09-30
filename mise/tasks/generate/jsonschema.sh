#!/usr/bin/env bash
#MISE description="Generate JSON Schema for Nais Apply"
set -euo pipefail

go run ./script/generate_jsonschema/
