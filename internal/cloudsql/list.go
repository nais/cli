package cloudsql

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
	"k8s.io/utils/ptr"
)

const consoleBaseURL = "https://console.nav.cloud.nais.io"

type Instance struct {
	Name             output.Link `json:"name"`
	Type             string      `json:"type"`
	Environment      string      `json:"environment"`
	Version          string      `heading:"Version" json:"version"`
	HighAvailability bool        `heading:"HA" json:"high_availability"`
	Audit            *bool       `json:"audit,omitempty"`
	State            State       `json:"state"`
}

type State string

func (s State) String() string {
	switch s {
	case State(gql.SqlInstanceStateRunnable):
		return "Runnable"
	case State(gql.SqlInstanceStateStopped):
		return "<error>Stopped</error>"
	case State(gql.SqlInstanceStateSuspended):
		return "<error>Suspended</error>"
	case State(gql.SqlInstanceStatePendingCreate):
		return "Pending Create"
	case State(gql.SqlInstanceStatePendingDelete):
		return "Pending Delete"
	case State(gql.SqlInstanceStateMaintenance):
		return "Maintenance"
	case State(gql.SqlInstanceStateFailed):
		return "<error>Failed</error>"
	}
	return "<info>Unknown</info>"
}

func GetTeamCloudSQLInstances(ctx context.Context, team string, environments []string, labelFilters []gql.LabelFilter) ([]Instance, error) {
	_ = `# @genqlient
		query GetTeamCloudSQLInstances($team: Slug!, $sqlFilter: SqlInstanceFilter) {
			team(slug: $team) {
				sqlInstances(first: 1000, filter: $sqlFilter) {
					nodes {
						name
						teamEnvironment { environment { name } }
						version
						highAvailability
						# @genqlient(pointer: true)
						auditLog { logUrl }
						state
					}
				}
			}
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := gql.GetTeamCloudSQLInstances(ctx, client, team, &gql.SqlInstanceFilter{Labels: labelFilters})
	if err != nil {
		return nil, err
	}
	return instancesFromTeam(resp.Team, team, environments), nil
}

func instancesFromTeam(teamData gql.GetTeamCloudSQLInstancesTeam, team string, environments []string) []Instance {
	var ret []Instance
	for _, s := range teamData.SqlInstances.Nodes {
		env := s.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}
		ret = append(ret, Instance{
			Name: output.Link{Name: s.Name, URL: fmt.Sprintf("%s/team/%s/%s/cloudsql/%s", consoleBaseURL, team, env, s.Name)},
			Type: "Cloud SQL", Environment: env, Version: ptr.Deref(s.Version, ""),
			HighAvailability: s.HighAvailability, Audit: ptr.To(s.AuditLog != nil), State: State(s.State),
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Name.Name == ret[j].Name.Name {
			return ret[i].Environment < ret[j].Environment
		}
		return ret[i].Name.Name < ret[j].Name.Name
	})
	return ret
}
