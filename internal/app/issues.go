package app

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ApplicationIssue struct {
	Type        gql.IssueType
	Severity    gql.Severity
	Message     string
	Environment string
}

func GetApplicationIssues(ctx context.Context, slug, name string, envs []string) ([]ApplicationIssue, error) {
	_ = `# @genqlient
		query GetApplicationIssues($slug: Slug!, $name: String!, $env: [String!]) {
		  team(slug: $slug) {
		    applications(filter: { name: $name, environments: $env }) {
		      nodes {
				teamEnvironment{
          		  environment {
              	    name
		 	      }
		 	    }
			    issues(first: 500) {
			      nodes {
					__typename
		            severity
		            message
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

	resp, err := gql.GetApplicationIssues(ctx, client, slug, name, envs)
	if err != nil {
		return nil, err
	}
	if len(resp.Team.Applications.Nodes) == 0 {
		return nil, nil
	}
	ret := make([]ApplicationIssue, 0)
	for _, app := range resp.Team.Applications.Nodes {
		for _, issue := range app.Issues.Nodes {
			ret = append(ret, ApplicationIssue{
				Type:        gql.IssueType(issue.GetTypename()),
				Severity:    issue.GetSeverity(),
				Message:     issue.GetMessage(),
				Environment: app.TeamEnvironment.Environment.Name,
			})
		}
	}
	return ret, nil
}
