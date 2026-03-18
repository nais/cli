package job

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func DeleteJobRun(ctx context.Context, team, environment, runName string) (bool, error) {
	_ = `# @genqlient
		mutation DeleteJobRun($team: Slug!, $env: String!, $runName: String!) {
			deleteJobRun(input: { teamSlug: $team, environmentName: $env, runName: $runName }) {
				success
			}
		}
	`

	if environment == "" {
		return false, fmt.Errorf("exactly one environment must be specified")
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}

	resp, err := gql.DeleteJobRun(ctx, client, team, environment, runName)
	if err != nil {
		return false, err
	}

	return resp.DeleteJobRun.Success, nil
}
