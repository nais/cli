package issues

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func GetAll(ctx context.Context, teamSlug string) ([]gql.GetAllIssuesTeamIssuesIssueConnectionNodesIssue, error) {
	_ = `# @genqlient
	query GetAllIssues($teamSlug: Slug!) {
	  team(slug: $teamSlug) {
		issues {
		  nodes {
			teamEnvironment {
			  environment {
				name
			  }
			}
			severity
			message
			__typename
		  }
		}
	  }
	}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetAllIssues(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Issues.Nodes, nil
}
