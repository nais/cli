package app

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/formatting"
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

type (
	LastUpdated time.Time
	State       string
)

type Application struct {
	State         State          `json:"state"`
	Name          string         `json:"name"`
	Environment   string         `json:"environment"`
	InstancesInfo *InstancesInfo `heading:"Running" json:"running"`
	IssueInfo     *IssueInfo     `heading:"Issues" json:"issue_info"`
	LastUpdated   LastUpdated    `heading:"Last Updated" json:"last_updated"`
}

func (s State) String() string {
	if s == State(gql.ApplicationStateRunning) {
		return "Running"
	} else if s == State(gql.ApplicationStateNotRunning) {
		return "<error>Not running</error>"
	}
	return "<info>Unknown</info>"
}

func (i IssueInfo) String() string {
	return formatting.ColoredSeverityString(fmt.Sprintf("%v %v", i.Count, i.Severity), gql.Severity(i.Severity))
}

func (i InstancesInfo) String() string {
	return fmt.Sprintf("%v/%v", i.Running, i.Total)
}

func (a LastUpdated) String() string {
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
		return fmt.Sprintf("%vs", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%vm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%vh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%vd", hours/24)
	}
	return fmt.Sprintf("%vy", int(d.Hours()/24/365))
}

func (a LastUpdated) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%q", time.Time(a).Format(time.RFC3339)), nil
}

func GetApplicationNames(ctx context.Context, team string) ([]string, error) {
	_ = `# @genqlient
		query GetApplicationNames($team: Slug!) {
		  team(slug: $team) {
			  applications(first: 1000) {
		      nodes {
		        name
		      }
		    }
		  }
		}
		`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetApplicationNames(ctx, client, team)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)

	for _, app := range resp.Team.Applications.Nodes {
		ret = append(ret, app.Name)
	}
	return ret, nil
}

func GetTeamApplications(ctx context.Context, team string, orderBy gql.ApplicationOrder, filter gql.TeamApplicationsFilter) ([]Application, error) {
	_ = `# @genqlient
		query GetTeamApplications($team: Slug!, $orderBy: ApplicationOrder, $filter: TeamApplicationsFilter) {
		  team(slug: $team) {
			  applications(orderBy: $orderBy, filter: $filter, first: 1000) {
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
		var lastUpdated LastUpdated
		if len(app.Deployments.GetNodes()) > 0 {
			lastUpdated = LastUpdated(app.Deployments.GetNodes()[0].GetCreatedAt())
		}

		ret = append(ret, Application{
			Name:          app.Name,
			Environment:   app.TeamEnvironment.Environment.Name,
			State:         State(app.State),
			LastUpdated:   lastUpdated,
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
