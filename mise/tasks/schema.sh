#!/usr/bin/env bash
#MISE description="Update GraphQL schema from local Nais API"
set -euo pipefail

NAIS_API_LOCAL_EMAIL="dev.usersen@example.com" NAIS_API_LOCAL_HOST="localhost:3000" go run . alpha api schema > schema.graphql
