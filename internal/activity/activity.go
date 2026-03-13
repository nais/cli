package activity

import (
	"context"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Entry struct {
	CreatedAt    time.Time `heading:"Created" json:"created_at"`
	Actor        string    `json:"actor"`
	Environment  string    `json:"environment"`
	ResourceType string    `heading:"Resource Type" json:"resource_type"`
	ResourceName string    `heading:"Resource Name" json:"resource_name"`
	Message      string    `json:"message"`
}

func List(ctx context.Context, team string, activityTypes []gql.ActivityLogActivityType, limit int) ([]Entry, error) {
	_ = `# @genqlient
		query GetTeamActivity($team: Slug!, $activityTypes: [ActivityLogActivityType!], $first: Int) {
			team(slug: $team) {
				activityLog(first: $first, filter: { activityTypes: $activityTypes }) {
					nodes {
						actor
						createdAt
						message
						environmentName
						resourceType
						resourceName
					}
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetTeamActivity(ctx, client, team, activityTypes, limit)
	if err != nil {
		return nil, err
	}

	ret := make([]Entry, 0, len(resp.Team.ActivityLog.Nodes))
	for _, entry := range resp.Team.ActivityLog.Nodes {
		env := entry.GetEnvironmentName()
		if env == "" {
			env = "N/A"
		}

		ret = append(ret, Entry{
			CreatedAt:    entry.GetCreatedAt(),
			Actor:        entry.GetActor(),
			Environment:  env,
			ResourceType: string(entry.GetResourceType()),
			ResourceName: entry.GetResourceName(),
			Message:      entry.GetMessage(),
		})
	}

	return ret, nil
}
