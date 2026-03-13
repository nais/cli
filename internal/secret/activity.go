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

type secretActivityEntry struct {
	CreatedAt       time.Time
	Actor           string
	Message         string
	EnvironmentName string
}

type secretActivityResource struct {
	Name           string
	DefaultEnvName string
	Entries        []secretActivityEntry
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

	resources := make([]secretActivityResource, 0, len(resp.Team.Secrets.Nodes))
	for _, s := range resp.Team.Secrets.Nodes {
		entries := make([]secretActivityEntry, 0, len(s.ActivityLog.Nodes))
		for _, entry := range s.ActivityLog.Nodes {
			entries = append(entries, secretActivityEntry{
				CreatedAt:       entry.GetCreatedAt(),
				Actor:           entry.GetActor(),
				Message:         entry.GetMessage(),
				EnvironmentName: entry.GetEnvironmentName(),
			})
		}

		resources = append(resources, secretActivityResource{
			Name:           s.Name,
			DefaultEnvName: s.TeamEnvironment.Environment.Name,
			Entries:        entries,
		})
	}

	ret, found := buildSecretActivity(resources, name, environments)
	return ret, found, nil
}

func buildSecretActivity(resources []secretActivityResource, name string, environments []string) ([]SecretActivity, bool) {
	found := false
	ret := make([]SecretActivity, 0)

	for _, s := range resources {
		if s.Name != name {
			continue
		}

		defaultEnv := s.DefaultEnvName
		if len(environments) > 0 && !slices.Contains(environments, defaultEnv) {
			continue
		}

		found = true

		for _, entry := range s.Entries {
			env := defaultEnv
			if entry.EnvironmentName != "" {
				env = entry.EnvironmentName
			}
			ret = append(ret, SecretActivity{
				CreatedAt:   entry.CreatedAt,
				Actor:       entry.Actor,
				Environment: env,
				Message:     entry.Message,
			})
		}
	}

	return ret, found
}
