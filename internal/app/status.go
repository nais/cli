package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type InstanceState string

func (s InstanceState) String() string {
	switch s {
	case "RUNNING":
		return "Running"
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
	Name     string        `json:"name"`
	State    InstanceState `json:"state"`
	Message  string        `json:"message,omitempty"`
	Restarts int           `json:"restarts"`
	Created  LastUpdated   `json:"created"`
}

type InstanceGroupStatus struct {
	Application string              `json:"application"`
	Environment string              `json:"environment"`
	Groups      []InstanceGroupInfo `json:"instanceGroups"`
}

type InstanceGroupInfo struct {
	Name             string         `json:"name"`
	Image            string         `json:"image"`
	Revision         int            `json:"revision"`
	ReadyInstances   int            `json:"readyInstances"`
	DesiredInstances int            `json:"desiredInstances"`
	Created          time.Time      `json:"created"`
	Current          bool           `json:"current"`
	Instances        []InstanceInfo `json:"instances"`
}

func GetApplicationStatus(ctx context.Context, slug, name string, envs []string) (*InstanceGroupStatus, error) {
	_ = `# @genqlient
		query GetApplicationStatus($slug: Slug!, $name: String!, $env: [String!]) {
		  team(slug: $slug) {
		    applications(filter: { name: $name, environments: $env }) {
		      nodes {
		        name
		        teamEnvironment {
		          environment {
		            name
		          }
		        }
		        instanceGroups {
		          name
		          image {
		            name
		            tag
		          }
		          revision
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

	resp, err := gql.GetApplicationStatus(ctx, client, slug, name, envs)
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Applications.Nodes) == 0 {
		return nil, fmt.Errorf("application %q not found", name)
	}

	app := resp.Team.Applications.Nodes[0]

	groups := make([]InstanceGroupInfo, 0, len(app.InstanceGroups))
	maxRevision := 0
	for _, ig := range app.InstanceGroups {
		if ig.Revision > maxRevision {
			maxRevision = ig.Revision
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
			})
		}

		groups = append(groups, InstanceGroupInfo{
			Name:             ig.Name,
			Image:            fmt.Sprintf("%s:%s", ig.Image.Name, ig.Image.Tag),
			Revision:         ig.Revision,
			ReadyInstances:   ig.ReadyInstances,
			DesiredInstances: ig.DesiredInstances,
			Created:          ig.Created,
			Current:          ig.Revision == maxRevision,
			Instances:        instances,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Revision > groups[j].Revision
	})

	return &InstanceGroupStatus{
		Application: app.Name,
		Environment: app.TeamEnvironment.Environment.Name,
		Groups:      groups,
	}, nil
}
