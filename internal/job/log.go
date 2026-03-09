package job

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func IsTerminalRunState(state gql.JobRunState) bool {
	switch state {
	case gql.JobRunStateSucceeded, gql.JobRunStateFailed:
		return true
	default:
		return false
	}
}

func GetLatestJobRunState(ctx context.Context, team, name, environment string) (gql.JobRunState, error) {
	_ = `# @genqlient
		query GetLatestJobRunState($team: Slug!, $name: String!, $env: [String!]) {
			team(slug: $team) {
				jobs(filter: { name: $name, environments: $env }) {
					nodes {
						runs(first: 1) {
							nodes {
								status {
									state
								}
							}
						}
					}
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return gql.JobRunStateUnknown, err
	}

	resp, err := gql.GetLatestJobRunState(ctx, client, team, name, []string{environment})
	if err != nil {
		return gql.JobRunStateUnknown, err
	}

	if len(resp.Team.Jobs.Nodes) == 0 {
		return gql.JobRunStateUnknown, nil
	}

	runs := resp.Team.Jobs.Nodes[0].Runs.Nodes
	if len(runs) == 0 {
		return gql.JobRunStateUnknown, nil
	}

	return runs[0].Status.State, nil
}
