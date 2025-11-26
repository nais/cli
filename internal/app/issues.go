package app

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ApplicationIssue struct {
	Total     int           `json:"total"`
	Type      gql.IssueType `json:"type"`
	Severity  gql.Severity  `json:"severity"`
	Message   string        `json:"message"`
	Ingresses []string      `json:"ingresses,omitempty"`
	RiskScore int           `json:"risk_score,omitempty"`
	Critical  int           `json:"critical,omitempty"`
}

func GetApplicationIssues(ctx context.Context, slug, name, env string) ([]ApplicationIssue, error) {
	_ = `# @genqlient
		query GetApplicationIssues($slug: Slug!, $name: String!, $env: String!) {
		  team(slug: $slug) {
		    applications(filter: { name: $name, environments: [$env] }) {
		      nodes {
		        issues(first: 500) {
		          pageInfo {
		            totalCount
		          }
		          nodes {
					__typename
		            severity
		            message
		            ... on DeprecatedIngressIssue {
		              ingresses
		            }
		            ... on VulnerableImageIssue{
		              riskScore
		              critical
		            }
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

	resp, err := gql.GetApplicationIssues(ctx, client, slug, name, env)
	if err != nil {
		return nil, err
	}
	ret := make([]ApplicationIssue, len(resp.Team.Applications.Nodes[0].Issues.Nodes))
	for i, issue := range resp.Team.Applications.Nodes {
		ret[i] = ApplicationIssue{
			Total:    issue.Issues.PageInfo.TotalCount,
			Type:     gql.IssueType(issue.Issues.Nodes[i].GetTypename()),
			Severity: issue.Issues.Nodes[i].GetSeverity(),
			Message:  issue.Issues.Nodes[i].GetMessage(),
		}
		switch c := issue.Issues.Nodes[0].(type) {
		case *gql.GetApplicationIssuesTeamApplicationsApplicationConnectionNodesApplicationIssuesIssueConnectionNodesDeprecatedIngressIssue:
			ret[i].Ingresses = c.Ingresses
		case *gql.GetApplicationIssuesTeamApplicationsApplicationConnectionNodesApplicationIssuesIssueConnectionNodesVulnerableImageIssue:
			ret[i].RiskScore = c.RiskScore
			ret[i].Critical = c.Critical
		}
	}
	return ret, nil
}
