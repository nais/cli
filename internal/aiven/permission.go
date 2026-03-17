package aiven

import (
	"slices"

	"github.com/nais/cli/internal/naisapi/gql"
)

func IsValidPermission(permission gql.CredentialPermission) bool {
	return slices.Contains(gql.AllCredentialPermission, permission)
}
