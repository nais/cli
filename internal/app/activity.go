package app

import (
	"context"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ApplicationActivity struct {
	CreatedAt   time.Time `heading:"Created" json:"created_at"`
	Actor       string    `json:"actor"`
	Environment string    `json:"environment"`
	Message     string    `json:"message"`
}

func GetApplicationActivity(ctx context.Context, team, name string, environments []string, activityTypes []gql.ActivityLogActivityType, limit int) ([]ApplicationActivity, bool, error) {
	_ = `# @genqlient
		query GetApplicationActivity($team: Slug!, $name: String!, $environments: [String!], $activityTypes: [ActivityLogActivityType!], $first: Int) {
			team(slug: $team) {
				applications(filter: { name: $name, environments: $environments }) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						activityLog(first: $first, filter: { activityTypes: $activityTypes }) {
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
		return nil, false, err
	}

	resp, err := gql.GetApplicationActivity(ctx, client, team, name, environments, activityTypes, limit)
	if err != nil {
		return nil, false, err
	}

	found := false

	ret := make([]ApplicationActivity, 0)
	for _, a := range resp.Team.Applications.Nodes {
		if a.Name != name {
			continue
		}

		found = true
		defaultEnv := a.TeamEnvironment.Environment.Name
		for _, entry := range a.ActivityLog.Nodes {
			env := defaultEnv
			if entry.GetEnvironmentName() != "" {
				env = entry.GetEnvironmentName()
			}
			ret = append(ret, ApplicationActivity{
				CreatedAt:   entry.GetCreatedAt(),
				Actor:       entry.GetActor(),
				Environment: env,
				Message:     entry.GetMessage(),
			})
		}
	}

	return ret, found, nil
}
