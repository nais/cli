#!/usr/bin/env bash
#MISE description="Update the graphql schema, using the live nais-api"
set -euo pipefail
go run . api schema >schema.graphql
