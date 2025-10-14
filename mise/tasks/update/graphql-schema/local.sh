#!/usr/bin/env bash
#MISE description="Update the graphql schema, using a locally running nais-api (localhost:3000)"
set -euo pipefail
NAIS_API_LOCAL_HOST="localhost:3000" NAIS_API_LOCAL_EMAIL="dev.usersen@example.com" go run . alpha api schema >schema.graphql
