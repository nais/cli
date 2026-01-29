package naisapi

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi/gql"
)

// SecretValue represents a key-value pair from a secret
type SecretValue struct {
	Name  string
	Value string
}

// ViewSecretValues retrieves the values of a secret. This requires team membership
// and a reason for access. The access is logged for auditing purposes.
func ViewSecretValues(ctx context.Context, team, environmentName, secretName, reason string) ([]SecretValue, error) {
	_ = `# @genqlient
mutation ViewSecretValues($input: ViewSecretValuesInput!) {
viewSecretValues(input: $input) {
values {
name
value
}
}
}
`

	client, err := GraphqlClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating GraphQL client: %w", err)
	}

	resp, err := gql.ViewSecretValues(ctx, client, gql.ViewSecretValuesInput{
		Name:        secretName,
		Environment: environmentName,
		Team:        team,
		Reason:      reason,
	})
	if err != nil {
		return nil, fmt.Errorf("viewing secret values: %w", err)
	}

	values := make([]SecretValue, len(resp.ViewSecretValues.Values))
	for i, v := range resp.ViewSecretValues.Values {
		values[i] = SecretValue{
			Name:  v.Name,
			Value: v.Value,
		}
	}

	return values, nil
}
