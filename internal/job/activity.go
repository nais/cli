package job

import (
	"context"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type JobActivity struct {
	CreatedAt   time.Time `heading:"Created" json:"created_at"`
	Actor       string    `json:"actor"`
	Environment string    `json:"environment"`
	Message     string    `json:"message"`
}

func GetJobActivity(ctx context.Context, team, name string, environments []string, limit int) ([]JobActivity, error) {
	_ = `# @genqlient
		query GetJobActivity($team: Slug!, $name: String!, $env: [String!], $first: Int) {
			team(slug: $team) {
				jobs(filter: { name: $name, environments: $env }) {
					nodes {
						teamEnvironment {
							environment {
								name
							}
						}
						activityLog(first: $first) {
							nodes {
								actor
								createdAt
								message
								environmentName
							}
						}
					}
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetJobActivity(ctx, client, team, name, environments, limit)
	if err != nil {
		return nil, err
	}

	ret := make([]JobActivity, 0)
	for _, j := range resp.Team.Jobs.Nodes {
		defaultEnv := j.TeamEnvironment.Environment.Name
		for _, entry := range j.ActivityLog.Nodes {
			env := defaultEnv
			if entry.GetEnvironmentName() != "" {
				env = entry.GetEnvironmentName()
			}
			ret = append(ret, JobActivity{
				CreatedAt:   entry.GetCreatedAt(),
				Actor:       entry.GetActor(),
				Environment: env,
				Message:     entry.GetMessage(),
			})
		}
	}

	return ret, nil
}
