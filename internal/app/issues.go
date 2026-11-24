package app

import (
	"context"

	"github.com/nais/cli/internal/formatting"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Severity gql.Severity

type ApplicationIssue struct {
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Environment string   `json:"environment"`
}

func (s Severity) String() string {
	return formatting.ColoredSeverityString(string(s), gql.Severity(s))
}

func GetApplicationIssues(ctx context.Context, slug, name string, environment string, filter *gql.ResourceIssueFilter) ([]ApplicationIssue, error) {
	_ = `# @genqlient
		query GetApplicationIssues($slug: Slug!, $name: String!, $env: [String!], $filter: ResourceIssueFilter) {
			team(slug: $slug) {
				applications(filter: { name: $name, environments: $env }) {
					nodes {
						name
						teamEnvironment { environment { name } }
						issues(first: 500, filter: $filter) {
							nodes { severity message }
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

	var environments []string
	if environment != "" {
		environments = []string{environment}
	}
	resp, err := gql.GetApplicationIssues(ctx, client, slug, name, environments, filter)
	if err != nil {
		return nil, err
	}
	ret := make([]ApplicationIssue, 0)
	for _, app := range resp.Team.Applications.Nodes {
		if app.Name != name {
			continue
		}
		for _, issue := range app.Issues.Nodes {
			ret = append(ret, ApplicationIssue{
				Severity:    Severity(issue.GetSeverity()),
				Message:     issue.GetMessage(),
				Environment: app.TeamEnvironment.Environment.Name,
			})
		}
	}
	return ret, nil
}
