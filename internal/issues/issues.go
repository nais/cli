package issues

import (
	"context"
	"fmt"

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
	ID           string
}

func GetAll(ctx context.Context, teamSlug string, issueFilter gql.IssueFilter) ([]Issue, error) {
	_ = `# @genqlient
		# @genqlient(for: "IssueFilter.issueType", omitempty: true)
		# @genqlient(for: "IssueFilter.severity", omitempty: true)
		# @genqlient(for: "IssueFilter.resourceType", omitempty: true)
		# @genqlient(for: "IssueFilter.resourceName", omitempty: true)
		# @genqlient(for: "IssueFilter.environments", omitempty: true)
	query GetAllIssues(
		$teamSlug: Slug!,
		$filter: IssueFilter
	) {
	  team(slug: $teamSlug) {
		  issues(filter: $filter first: 999) {
		  nodes {
			teamEnvironment {
			  environment {
				name
			  }
			}
			id
			severity
			message
			__typename
			... on DeprecatedIngressIssue {
			  application {
				name
				__typename
			  }
			}
			... on DeprecatedRegistryIssue {
			  workload {
				name
				__typename
			  }
			}
			... on FailedSynchronizationIssue {
			  workload {
				name
				__typename
			  }
			}
			... on InvalidSpecIssue {
			  workload {
				name
				__typename
			  }
			}
			... on LastRunFailedIssue {
			  job {
				name
				__typename
			  }
			}
			... on MissingSbomIssue {
			  workload {
				name
				__typename
			  }
			}
			... on NoRunningInstancesIssue {
			  workload {
				name
				__typename
			  }
			}
			... on OpenSearchIssue {
			  openSearch {
				name
				__typename
			  }
			}
			... on SqlInstanceStateIssue {
			  sqlInstance {
				name
				__typename
			  }
			}
			... on SqlInstanceVersionIssue {
			  sqlInstance {
				name
				__typename
			  }
			}
			... on ValkeyIssue {
			  valkey {
				name
				__typename
			  }
			}
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

	resp, err := gql.GetAllIssues(ctx, client, teamSlug, issueFilter)
	if err != nil {
		return nil, fmt.Errorf("graphql: %w", err)
	}

	ret := make([]Issue, 0)

	for _, issue := range resp.Team.Issues.Nodes {
		i := Issue{
			ID:          issue.GetId(),
			Environment: issue.GetTeamEnvironment().Environment.Name,
			Severity:    string(issue.GetSeverity()),
			Message:     issue.GetMessage(),
			IssueType:   issue.GetTypename(),
		}
		switch c := issue.(type) {
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesDeprecatedIngressIssue:
			i.ResourceName = c.Application.GetName()
			i.ResourceType = c.Application.GetTypename()
			i.IssueType = "Deprecated ingress"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesDeprecatedRegistryIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "Deprecated registry"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesFailedSynchronizationIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "Failed synchronization"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesInvalidSpecIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "Invalid spec"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesLastRunFailedIssue:
			i.ResourceName = c.Job.GetName()
			i.ResourceType = c.Job.GetTypename()
			i.IssueType = "Last job run failed"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesMissingSbomIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "Missing SBOM"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesNoRunningInstancesIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "No running instances"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesOpenSearchIssue:
			i.ResourceName = c.OpenSearch.GetName()
			i.ResourceType = c.OpenSearch.GetTypename()
			i.IssueType = "Opensearch issue"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesSqlInstanceStateIssue:
			i.ResourceName = c.SqlInstance.GetName()
			i.ResourceType = c.SqlInstance.GetTypename()
			i.IssueType = "SQL instance problem"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesSqlInstanceVersionIssue:
			i.ResourceName = c.SqlInstance.GetName()
			i.ResourceType = c.SqlInstance.GetTypename()
			i.IssueType = "SQL instance version outdated"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesValkeyIssue:
			i.ResourceName = c.Valkey.GetName()
			i.ResourceType = c.Valkey.GetTypename()
			i.IssueType = "Valkey issue"
		case *gql.GetAllIssuesTeamIssuesIssueConnectionNodesVulnerableImageIssue:
			i.ResourceName = c.GetWorkload().GetName()
			i.ResourceType = c.GetWorkload().GetTypename()
			i.IssueType = "Vulnerable image"
		}
		ret = append(ret, i)

	}
	return ret, nil
}
