package secret

import (
	"context"
	"slices"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type SecretActivity struct {
	CreatedAt   time.Time `heading:"Created" json:"created_at"`
	Actor       string    `json:"actor"`
	Environment string    `json:"environment"`
	Message     string    `json:"message"`
}

func GetActivity(ctx context.Context, team, name string, environments []string, activityTypes []gql.ActivityLogActivityType, limit int) ([]SecretActivity, bool, error) {
	_ = `# @genqlient
		query GetSecretActivity($team: Slug!, $name: String!, $activityTypes: [ActivityLogActivityType!], $first: Int) {
			team(slug: $team) {
				secrets(filter: { name: $name }, first: 1000) {
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

	resp, err := gql.GetSecretActivity(ctx, client, team, name, activityTypes, limit)
	if err != nil {
		return nil, false, err
	}

	found := false

	ret := make([]SecretActivity, 0)
	for _, s := range resp.Team.Secrets.Nodes {
		if s.Name != name {
			continue
		}

		found = true
		defaultEnv := s.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, defaultEnv) {
			continue
		}

		for _, entry := range s.ActivityLog.Nodes {
			env := defaultEnv
			if entry.GetEnvironmentName() != "" {
				env = entry.GetEnvironmentName()
			}
			ret = append(ret, SecretActivity{
				CreatedAt:   entry.GetCreatedAt(),
				Actor:       entry.GetActor(),
				Environment: env,
				Message:     entry.GetMessage(),
			})
		}
	}

	return ret, found, nil
}
