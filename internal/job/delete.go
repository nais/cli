package job

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func DeleteJobRun(ctx context.Context, team, environment, runName string) error {
	_ = `# @genqlient
		mutation DeleteJobRun($team: Slug!, $env: String!, $runName: String!) {
			deleteJobRun(input: { teamSlug: $team, environmentName: $env, runName: $runName }) {
				success
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return err
	}

	_, err = gql.DeleteJobRun(ctx, client, team, environment, runName)
	return err
}
