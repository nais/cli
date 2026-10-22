package postgres

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// GetPostgres returns configured metadata and branch observations, not access credentials.
func GetPostgres(ctx context.Context, team, environment, name string) (*gql.GetPostgresTeamEnvironmentPostgres, error) {
	_ = `# @genqlient
	query GetPostgres($team: Slug!, $environment: String!, $postgres: String!) {
		team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) {
			name majorVersion highAvailability
			resources { cpu memory diskSize }
			labels { key value }
			desiredActiveBranch activeBranch { name }
			branches(first: 1000) { nodes { name state } }
		} } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := gql.GetPostgres(ctx, client, team, environment, name)
	if err != nil {
		return nil, fmt.Errorf("fetching Postgres %q: %w", name, err)
	}
	return &result.Team.Environment.Postgres, nil
}
