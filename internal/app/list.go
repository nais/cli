package app

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type IssueInfo struct {
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}

type InstancesInfo struct {
	Total   int `json:"total"`
	Running int `json:"running"`
}

type Age time.Time

type Application struct {
	Name          string         `json:"name"`
	Environment   string         `json:"environment"`
	InstancesInfo *InstancesInfo `heading:"Running" json:"running"`
	State         string
	IssueInfo     *IssueInfo `heading:"Issues" json:"issue_info"`
	Age           Age        `json:"age"`
}

func (i IssueInfo) String() string {
	return fmt.Sprintf("%v:%v", i.Severity, i.Count)
}

func (i InstancesInfo) String() string {
	return fmt.Sprintf("%v/%v", i.Running, i.Total)
}

func (a Age) String() string {
	t := time.Time(a)
	if t.IsZero() {
		return "<unknown>"
	}

	d := time.Since(time.Time(a))
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func GetTeamApplications(ctx context.Context, team string, orderBy gql.ApplicationOrder, filter gql.TeamApplicationsFilter) ([]Application, error) {
	_ = `# @genqlient
		query GetTeamApplications($team: Slug!, $orderBy: ApplicationOrder, $filter: TeamApplicationsFilter) {
		  team(slug: $team) {
		    applications(orderBy: $orderBy, filter: $filter) {
		      nodes {
		        name
		        teamEnvironment {
		          environment {
		            name
		          }
		        }
		        state
		        issues {
		          nodes {
		            severity
		          }
		        }
		        deployments(first: 1) {
		          nodes {
		            createdAt
		          }
		        }
		        instances {
		          nodes {
		            status {
		              state
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

	resp, err := gql.GetTeamApplications(ctx, client, team, orderBy, filter)
	if err != nil {
		return nil, err
	}
	ret := make([]Application, 0)

	for _, app := range resp.Team.Applications.Nodes {
		var lastUpdated Age
		if len(app.Deployments.GetNodes()) > 0 {
			lastUpdated = Age(app.Deployments.GetNodes()[0].GetCreatedAt())
		}

		ret = append(ret, Application{
			Name:          app.Name,
			Environment:   app.TeamEnvironment.Environment.Name,
			State:         string(app.State),
			Age:           lastUpdated,
			IssueInfo:     issueInfo(app.Issues.GetNodes()),
			InstancesInfo: instanceInfo(app.Instances.GetNodes()),
		})
	}
	return ret, nil
}

func instanceInfo(instances []gql.GetTeamApplicationsTeamApplicationsApplicationConnectionNodesApplicationInstancesApplicationInstanceConnectionNodesApplicationInstance) *InstancesInfo {
	running := 0
	for _, instance := range instances {
		if instance.Status.State == gql.ApplicationInstanceStateRunning {
			running++
		}
	}
	return &InstancesInfo{
		Total:   len(instances),
		Running: running,
	}
}

func age(deployments []gql.GetTeamApplicationsTeamApplicationsApplicationConnectionNodesApplicationDeploymentsDeploymentConnectionNodesDeployment) *time.Duration {
	if len(deployments) == 0 {
		return nil
	}

	ret := time.Since(deployments[0].GetCreatedAt())
	return &ret
}

func issueInfo(issues []gql.GetTeamApplicationsTeamApplicationsApplicationConnectionNodesApplicationIssuesIssueConnectionNodesIssue) *IssueInfo {
	if len(issues) == 0 {
		return nil
	}
	counts := map[gql.Severity]int{
		gql.SeverityCritical: 0,
		gql.SeverityWarning:  0,
		gql.SeverityTodo:     0,
	}

	for _, issue := range issues {
		counts[issue.GetSeverity()]++
	}

	if counts[gql.SeverityCritical] > 0 {
		return &IssueInfo{
			Severity: string(gql.SeverityCritical),
			Count:    counts[gql.SeverityCritical],
		}
	}
	if counts[gql.SeverityWarning] > 0 {
		return &IssueInfo{
			Severity: string(gql.SeverityWarning),
			Count:    counts[gql.SeverityWarning],
		}
	}
	return &IssueInfo{
		Severity: string(gql.SeverityTodo),
		Count:    counts[gql.SeverityTodo],
	}
}
