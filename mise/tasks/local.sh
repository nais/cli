#!/usr/bin/env bash
#MISE description="Run Nais CLI locally, using a locally running nais-api (localhost:3000)"
set -euo pipefail
NAIS_API_LOCAL_HOST="localhost:3000" NAIS_API_LOCAL_EMAIL="dev.usersen@example.com" go run . $@
