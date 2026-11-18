package postgres

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// GetCloneSource returns the configuration a clone inherits from, and the earliest
// point in time covered by its backups.
func GetCloneSource(ctx context.Context, team, environment, name string) (*gql.GetCloneSourceTeamEnvironmentPostgres, error) {
	_ = `# @genqlient
	query GetCloneSource($team: Slug!, $environment: String!, $postgres: String!) {
		team(slug: $team) { environment(name: $environment) { postgres(name: $postgres) {
			name majorVersion highAvailability state earliestRestoreTime
			resources { cpu memory diskSize }
		} } }
	}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := gql.GetCloneSource(ctx, client, team, environment, name)
	if err != nil {
		return nil, fmt.Errorf("fetching Postgres %q: %w", name, err)
	}
	return &result.Team.Environment.Postgres, nil
}
