package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/workload"
)

type InstanceState string

func (s InstanceState) String() string {
	switch s {
	case "RUNNING":
		return "Running"
	case "PENDING":
		return "Waiting for placement"
	case "STARTING":
		return "<info>Starting</info>"
	case "FAILING":
		return "<error>Failing</error>"
	case "TERMINATED":
		return "Terminated"
	default:
		return "<info>Unknown</info>"
	}
}

type InstanceInfo struct {
	Name     string                 `json:"name"`
	State    InstanceState          `json:"state"`
	Message  string                 `json:"message,omitempty"`
	Restarts int                    `json:"restarts"`
	Created  LastUpdated            `json:"created"`
	Ready    bool                   `json:"ready"`
	LastExit *workload.InstanceExit `json:"lastExit,omitempty"`
}

type InstanceGroupStatus struct {
	Application string              `json:"application"`
	Environment string              `json:"environment"`
	Groups      []InstanceGroupInfo `json:"instanceGroups"`
	Rollout     *RolloutWarning     `json:"rollout,omitempty"`
	Scaling     *ScalingStatus      `json:"scaling,omitempty"`
}

type InstanceGroupInfo struct {
	Name             string         `json:"name"`
	Image            string         `json:"image"`
	ReadyInstances   int            `json:"readyInstances"`
	DesiredInstances int            `json:"desiredInstances"`
	Created          time.Time      `json:"created"`
	Current          bool           `json:"current"`
	Instances        []InstanceInfo `json:"instances"`
}

type RolloutWarning struct {
	Type     gql.IssueType                     `json:"type"`
	Severity gql.Severity                      `json:"severity"`
	Reason   gql.ApplicationRolloutStuckReason `json:"reason"`
	Message  string                            `json:"message"`
}

type ScalingStatus struct {
	State            gql.ApplicationScalingState `json:"state"`
	Message          string                      `json:"message"`
	CurrentInstances int                         `json:"currentInstances"`
	DesiredInstances int                         `json:"desiredInstances"`
}

func GetApplicationStatus(ctx context.Context, slug, name, env string) (*InstanceGroupStatus, error) {
	_ = `# @genqlient
        query GetApplicationStatus($slug: Slug!, $name: String!, $env: String!) {
          team(slug: $slug) {
            environment(name: $env) {
              application(name: $name) {
                name
                issues(first: 1, filter: { issueType: APPLICATION_ROLLOUT_STUCK }) {
                  nodes {
                    ... on ApplicationRolloutStuckIssue {
                      severity
                      reason
                      message
                    }
                  }
                }
                resources {
                  scaling {
                    status {
                      state
                      message
                      currentInstances
                      desiredInstances
                    }
                  }
                }
                instanceGroups {
                  name
                  image {
                    name
                    tag
                  }
                  created
                  readyInstances
                  desiredInstances
                  instances {
                    name
                    restarts
                    created
                    status {
                      state
                      message
                      ready
                      lastExitReason
                      lastExitCode
                      lastExitTimestamp
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

	resp, err := gql.GetApplicationStatus(ctx, client, slug, name, env)
	if err != nil {
		return nil, err
	}

	app := resp.Team.Environment.Application
	ret := &InstanceGroupStatus{Application: app.Name, Environment: env}
	for _, node := range app.Issues.Nodes {
		if issue, ok := node.(*gql.GetApplicationStatusTeamEnvironmentApplicationIssuesIssueConnectionNodesApplicationRolloutStuckIssue); ok {
			ret.Rollout = &RolloutWarning{
				Type: gql.IssueTypeApplicationRolloutStuck, Severity: issue.Severity,
				Reason: issue.Reason, Message: issue.Message,
			}
		}
	}
	if status := app.Resources.Scaling.Status; status != nil {
		ret.Scaling = &ScalingStatus{
			State: status.State, Message: status.Message,
			CurrentInstances: status.CurrentInstances, DesiredInstances: status.DesiredInstances,
		}
	}

	for _, ig := range app.InstanceGroups {
		instances := make([]InstanceInfo, 0, len(ig.Instances))
		for _, inst := range ig.Instances {
			instances = append(instances, InstanceInfo{
				Name:     inst.Name,
				State:    InstanceState(inst.Status.State),
				Message:  inst.Status.Message,
				Restarts: inst.Restarts,
				Created:  LastUpdated(inst.Created),
				Ready:    inst.Status.Ready,
				LastExit: workload.LastExit(inst.Status.LastExitReason, inst.Status.LastExitCode, inst.Status.LastExitTimestamp),
			})
		}

		ret.Groups = append(ret.Groups, InstanceGroupInfo{
			Name:             ig.Name,
			Image:            fmt.Sprintf("%s:%s", ig.Image.Name, ig.Image.Tag),
			ReadyInstances:   ig.ReadyInstances,
			DesiredInstances: ig.DesiredInstances,
			Created:          ig.Created,
			Instances:        instances,
		})
	}

	// Keep the legacy JSON current flag: it marks the newest observed group, not which group serves traffic.
	sort.SliceStable(ret.Groups, func(i, j int) bool {
		return ret.Groups[i].Created.After(ret.Groups[j].Created)
	})
	if len(ret.Groups) > 0 {
		ret.Groups[0].Current = true
	}
	return ret, nil
}
