package issues

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Issue struct {
	Environment  string
	Severity     string
	Message      string
	ResourceName string
	ResourceType string
	IssueType    string
}

func GetAll(ctx context.Context, teamSlug string) ([]Issue, error) {
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
			... on VulnerableImageIssue {
			  workload {
				name
				__typename
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

	resp, err := gql.GetAllIssues(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	ret := make([]Issue, 0)

	for _, issue := range resp.Team.Issues.Nodes {
		i := Issue{
			Environment: issue.GetTeamEnvironment().Environment.Name,
			Severity:    string(issue.GetSeverity()),
			Message:     issue.GetMessage(),
			IssueType:   issue.GetTypename(),
		}
		switch c := issue.(type) {
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesVulnerableImageIssue:
			i.ResourceName = c.Workload.GetName()
			i.ResourceType = c.Workload.GetTypename()
		}
		ret = append(ret, i)

	}
	return ret, nil
}
