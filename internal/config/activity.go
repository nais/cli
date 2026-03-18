package config

import (
	"context"
	"slices"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ConfigActivity struct {
	CreatedAt   time.Time `heading:"Created" json:"created_at"`
	Actor       string    `json:"actor"`
	Environment string    `json:"environment"`
	Message     string    `json:"message"`
}

type configActivityEntry struct {
	CreatedAt       time.Time
	Actor           string
	Message         string
	EnvironmentName string
}

type configActivityResource struct {
	Name           string
	DefaultEnvName string
	Entries        []configActivityEntry
}

func GetActivity(ctx context.Context, team, name string, environments []string, activityTypes []gql.ActivityLogActivityType, limit int) ([]ConfigActivity, bool, error) {
	_ = `# @genqlient
		query GetConfigActivity($team: Slug!, $name: String!, $activityTypes: [ActivityLogActivityType!], $first: Int) {
			team(slug: $team) {
				configs(filter: { name: $name }, first: 1000) {
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

	resp, err := gql.GetConfigActivity(ctx, client, team, name, activityTypes, limit)
	if err != nil {
		return nil, false, err
	}

	resources := make([]configActivityResource, 0, len(resp.Team.Configs.Nodes))
	for _, c := range resp.Team.Configs.Nodes {
		entries := make([]configActivityEntry, 0, len(c.ActivityLog.Nodes))
		for _, entry := range c.ActivityLog.Nodes {
			entries = append(entries, configActivityEntry{
				CreatedAt:       entry.GetCreatedAt(),
				Actor:           entry.GetActor(),
				Message:         entry.GetMessage(),
				EnvironmentName: entry.GetEnvironmentName(),
			})
		}

		resources = append(resources, configActivityResource{
			Name:           c.Name,
			DefaultEnvName: c.TeamEnvironment.Environment.Name,
			Entries:        entries,
		})
	}

	ret, found := buildConfigActivity(resources, name, environments)
	return ret, found, nil
}

func buildConfigActivity(resources []configActivityResource, name string, environments []string) ([]ConfigActivity, bool) {
	found := false
	ret := make([]ConfigActivity, 0)

	for _, c := range resources {
		if c.Name != name {
			continue
		}

		defaultEnv := c.DefaultEnvName
		if len(environments) > 0 && !slices.Contains(environments, defaultEnv) {
			continue
		}

		found = true

		for _, entry := range c.Entries {
			env := defaultEnv
			if entry.EnvironmentName != "" {
				env = entry.EnvironmentName
			}
			ret = append(ret, ConfigActivity{
				CreatedAt:   entry.CreatedAt,
				Actor:       entry.Actor,
				Environment: env,
				Message:     entry.Message,
			})
		}
	}

	return ret, found
}
