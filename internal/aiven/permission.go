package aiven

import (
	"slices"

	"github.com/nais/cli/internal/naisapi/gql"
)

func IsValidPermission(permission gql.AivenPermission) bool {
	return slices.Contains(gql.AllAivenPermission, permission)
}
