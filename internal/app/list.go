package app

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Application struct {
	Name            string `json:"name"`
	Environment     string `json:"environment"`
	State           string `json:"state"`
	Vulnerabilities int    `json:"vulnerabilities"`
	Issues          int    `heading:"Critical Issues" json:"issues"`
}

func GetTeamApplications(ctx context.Context, teamSlug string) ([]Application, error) {
	_ = `# @genqlient
		query GetTeamApplications($slug: Slug!) {
		  team(slug: $slug) {
		    applications(first: 500) {
		      nodes {
		        name
		        state
		        totalIssues: issues {
		          pageInfo {
		            totalCount
		          }
		        }
		        image {
		          vulnerabilitySummary {
		            total
		          }
		        }
		        teamEnvironment {
		          environment {
		            name
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

	resp, err := gql.GetTeamApplications(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}
	ret := make([]Application, len(resp.Team.Applications.Nodes))
	for i, app := range resp.Team.Applications.Nodes {
		ret[i] = Application{
			Name:            app.Name,
			Environment:     app.TeamEnvironment.Environment.Name,
			State:           string(app.State),
			Vulnerabilities: app.Image.VulnerabilitySummary.Total,
			Issues:          app.TotalIssues.PageInfo.TotalCount,
		}
	}
	return ret, nil
}
