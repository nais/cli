package postgres

import (
	"context"
	"slices"
	"sort"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type PostgresInstance struct {
	Name             string `json:"name"`
	Environment      string `json:"environment"`
	Version          string `heading:"Version" json:"version"`
	HighAvailability bool   `heading:"HA" json:"high_availability"`
	State            State  `json:"state"`
}

type State string

func (s State) String() string {
	switch s {
	case State(gql.PostgresInstanceStateAvailable):
		return "Available"
	case State(gql.PostgresInstanceStateProgressing):
		return "Progressing"
	case State(gql.PostgresInstanceStateDegraded):
		return "<error>Degraded</error>"
	default:
		return "<info>Unknown</info>"
	}
}

func GetTeamPostgresInstances(ctx context.Context, team string, environments []string) ([]PostgresInstance, error) {
	_ = `# @genqlient
		query GetTeamPostgresInstances($team: Slug!, $orderBy: PostgresInstanceOrder) {
			team(slug: $team) {
				postgresInstances(first: 1000, orderBy: $orderBy) {
					nodes {
						name
						teamEnvironment {
							environment {
								name
							}
						}
						majorVersion
						highAvailability
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

	resp, err := gql.GetTeamPostgresInstances(ctx, client, team, gql.PostgresInstanceOrder{
		Field:     gql.PostgresInstanceOrderFieldName,
		Direction: gql.OrderDirectionAsc,
	})
	if err != nil {
		return nil, err
	}

	ret := make([]PostgresInstance, 0)
	for _, p := range resp.Team.PostgresInstances.Nodes {
		env := p.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}

		ret = append(ret, PostgresInstance{
			Name:             p.Name,
			Environment:      env,
			Version:          p.MajorVersion,
			HighAvailability: p.HighAvailability,
			State:            State(p.State),
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Environment == ret[j].Environment {
			return ret[i].Name < ret[j].Name
		}
		return ret[i].Environment < ret[j].Environment
	})

	return ret, nil
}
