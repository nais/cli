package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func GetApplicationInstances(ctx context.Context, team, app, env string) ([]string, error) {
	_ = `# @genqlient
		query GetApplicationInstances($team: Slug!, $orderBy: ApplicationOrder, $filter: TeamApplicationsFilter) {
		  team(slug: $team) {
			  applications(orderBy: $orderBy, filter: $filter, first: 1000) {
		      nodes {
		        instances {
		          nodes {
					name
		          }
		        }
		      }
		    }
		  }
		}
		`
	if env == "" {
		return nil, fmt.Errorf("environment must be specified to get application instances")
	}
	if app == "" {
		return nil, fmt.Errorf("application name must be specified to get application instances")
	}
	if team == "" {
		return nil, fmt.Errorf("team must be specified to get application instances")
	}

	filter := gql.TeamApplicationsFilter{
		Environments: []string{env},
		Name:         app,
	}
	orderBy := gql.ApplicationOrder{
		Field:     gql.ApplicationOrderFieldIssues,
		Direction: gql.OrderDirectionDesc,
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetApplicationInstances(ctx, client, team, orderBy, filter)
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Applications.Nodes) == 0 {
		return []string{}, nil
	}

	if len(resp.Team.Applications.Nodes) > 1 {
		return nil, fmt.Errorf("expected one application, got %d", len(resp.Team.Applications.Nodes))
	}

	ret := make([]string, 0)

	for _, i := range resp.Team.Applications.Nodes[0].Instances.Nodes {
		ret = append(ret, i.GetName())
	}
	return ret, nil
}
