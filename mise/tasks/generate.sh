#!/usr/bin/env bash
#MISE description="Generate GraphQL client code"
set -euo pipefail

go tool github.com/Khan/genqlient
