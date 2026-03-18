package postgres

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix/output"
)

const consoleBaseURL = "https://console.nav.cloud.nais.io"

type Instance struct {
	Name             output.Link `json:"name"`
	Type             string      `json:"type"`
	Environment      string      `json:"environment"`
	Version          string      `heading:"Version" json:"version"`
	HighAvailability bool        `heading:"HA" json:"high_availability"`
	Audit            bool        `json:"audit"`
	State            State       `json:"state"`
}

type State string

func (s State) String() string {
	// PostgresInstance states
	switch s {
	case State(gql.PostgresInstanceStateAvailable):
		return "Available"
	case State(gql.PostgresInstanceStateProgressing):
		return "Progressing"
	case State(gql.PostgresInstanceStateDegraded):
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

func GetTeamPostgresInstances(ctx context.Context, team string, environments []string) ([]Instance, error) {
	_ = `# @genqlient
		query GetTeamPostgresInstances($team: Slug!) {
			team(slug: $team) {
				postgresInstances(first: 1000) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						majorVersion
						highAvailability
						audit {
							enabled
						}
						state
					}
				}
				sqlInstances(first: 1000) {
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

	resp, err := gql.GetTeamPostgresInstances(ctx, client, team)
	if err != nil {
		return nil, err
	}

	var ret []Instance

	for _, p := range resp.Team.PostgresInstances.Nodes {
		env := p.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		ret = append(ret, Instance{
			Name: output.Link{
				Name: p.Name,
				URL:  fmt.Sprintf("%s/team/%s/%s/postgres/%s", consoleBaseURL, team, env, p.Name),
			},
			Type:             "PostgreSQL",
			Environment:      env,
			Version:          p.MajorVersion,
			HighAvailability: p.HighAvailability,
			Audit:            p.Audit.Enabled,
			State:            State(p.State),
		})
	}

	for _, s := range resp.Team.SqlInstances.Nodes {
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
			Version:          s.Version,
			HighAvailability: s.HighAvailability,
			Audit:            s.AuditLog != nil,
			State:            State(s.State),
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Name.Name == ret[j].Name.Name {
			return ret[i].Environment < ret[j].Environment
		}
		return ret[i].Name.Name < ret[j].Name.Name
	})

	return ret, nil
}
