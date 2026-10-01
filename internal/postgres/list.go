package postgres

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
	// PostgresBranch states
	switch s {
	case State(gql.PostgresBranchStateAvailable):
		return "Available"
	case State(gql.PostgresBranchStateProgressing):
		return "Progressing"
	case State(gql.PostgresBranchStateDegraded):
		return "<error>Degraded</error>"
	}

	// SqlInstance states
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

func GetTeamPostgresBranches(ctx context.Context, team string, environments []string, labelFilters []gql.LabelFilter) ([]Instance, error) {
	_ = `# @genqlient
		query GetTeamPostgresBranches($team: Slug!, $postgresFilter: PostgresBranchFilter, $sqlFilter: SqlInstanceFilter) {
			team(slug: $team) {
				postgresBranches(first: 1000, filter: $postgresFilter) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						postgres {
							name
							majorVersion
							highAvailability
						}
						state
					}
				}
				sqlInstances(first: 1000, filter: $sqlFilter) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						version
						highAvailability
						# @genqlient(pointer: true)
						auditLog {
							logUrl
						}
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

	postgresFilter := gql.PostgresBranchFilter{
		Environments: environments,
		Labels:       labelFilters,
	}
	sqlFilter := gql.SqlInstanceFilter{
		Labels: labelFilters,
	}

	resp, err := gql.GetTeamPostgresBranches(ctx, client, team, new(postgresFilter), new(sqlFilter))
	if err != nil {
		return nil, err
	}

	return instancesFromTeam(resp.Team, team, environments), nil
}

func instancesFromTeam(teamData gql.GetTeamPostgresBranchesTeam, team string, environments []string) []Instance {
	var ret []Instance

	for _, p := range teamData.PostgresBranches.Nodes {
		env := p.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		ret = append(ret, Instance{
			Name: output.Link{
				Name: p.Postgres.Name + "/" + p.Name,
				URL:  fmt.Sprintf("%s/team/%s/%s/postgres/%s", consoleBaseURL, team, env, p.Postgres.Name),
			},
			Type:             "PostgreSQL",
			Environment:      env,
			Version:          p.Postgres.MajorVersion,
			HighAvailability: p.Postgres.HighAvailability,
			State:            State(p.State),
		})
	}

	for _, s := range teamData.SqlInstances.Nodes {
		env := s.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		ret = append(ret, Instance{
			Name: output.Link{
				Name: s.Name,
				URL:  fmt.Sprintf("%s/team/%s/%s/cloudsql/%s", consoleBaseURL, team, env, s.Name),
			},
			Type:             "Cloud SQL",
			Environment:      env,
			Version:          ptr.Deref(s.Version, ""),
			HighAvailability: s.HighAvailability,
			Audit:            new(s.AuditLog != nil),
			State:            State(s.State),
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
