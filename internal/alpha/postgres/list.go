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
	State            State       `json:"state"`
}

type State string

func (s State) String() string {
	switch s {
	case State(gql.PostgresBranchStateAvailable):
		return "Available"
	case State(gql.PostgresBranchStateProgressing):
		return "Progressing"
	case State(gql.PostgresBranchStateDegraded):
		return "<error>Degraded</error>"
	}
	return "<info>Unknown</info>"
}

func GetTeamPostgresBranches(ctx context.Context, team string, environments []string, labelFilters []gql.LabelFilter) ([]Instance, error) {
	_ = `# @genqlient
		query GetTeamPostgresBranchesAlpha($team: Slug!, $postgresFilter: PostgresBranchFilter) {
			team(slug: $team) {
				postgresBranches(first: 1000, filter: $postgresFilter) {
					nodes {
						name
						teamEnvironment { environment { name } }
						postgres { name majorVersion highAvailability }
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
	resp, err := gql.GetTeamPostgresBranchesAlpha(ctx, client, team, &gql.PostgresBranchFilter{Environments: environments, Labels: labelFilters})
	if err != nil {
		return nil, err
	}
	return instancesFromTeam(resp.Team, team, environments), nil
}

func instancesFromTeam(data gql.GetTeamPostgresBranchesAlphaTeam, team string, environments []string) []Instance {
	var ret []Instance
	for _, p := range data.PostgresBranches.Nodes {
		env := p.TeamEnvironment.Environment.Name
		if len(environments) > 0 && !slices.Contains(environments, env) {
			continue
		}
		ret = append(ret, Instance{
			Name: output.Link{Name: p.Postgres.Name + "/" + p.Name, URL: fmt.Sprintf("%s/team/%s/%s/postgres/%s", consoleBaseURL, team, env, p.Postgres.Name)},
			Type: "PostgreSQL", Environment: env, Version: p.Postgres.MajorVersion,
			HighAvailability: p.Postgres.HighAvailability, State: State(p.State),
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
