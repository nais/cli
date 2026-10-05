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

type PostgresInstance struct {
	Name             output.Link `json:"name"`
	Environment      string      `json:"environment"`
	Version          string      `heading:"Version" json:"version"`
	HighAvailability bool        `heading:"HA" json:"high_availability"`
	ActiveBranch     string      `heading:"Active branch" json:"active_branch"`
}

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

func GetTeamPostgreses(ctx context.Context, team string, environments []string, labelFilters []gql.LabelFilter) ([]PostgresInstance, error) {
	_ = `# @genqlient
		query GetTeamPostgreses($team: Slug!, $after: Cursor, $filter: PostgresFilter) {
			team(slug: $team) {
				postgreses(first: 100, after: $after, filter: $filter) {
					nodes { name teamEnvironment { environment { name } } majorVersion highAvailability activeBranch { name } }
					pageInfo { hasNextPage endCursor }
				}
			}
		}
	`
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	filter := &gql.PostgresFilter{Environments: environments, Labels: labelFilters}
	var ret []PostgresInstance
	var after *string
	for {
		resp, err := gql.GetTeamPostgreses(ctx, client, team, after, filter)
		if err != nil {
			return nil, err
		}
		ret = append(ret, postgresesFromTeam(resp.Team, team)...)
		if !resp.Team.Postgreses.PageInfo.HasNextPage {
			break
		}
		next := resp.Team.Postgreses.PageInfo.EndCursor
		if next == nil || after != nil && *next == *after {
			return nil, fmt.Errorf("postgres list has another page but no new cursor")
		}
		after = next
	}
	return ret, nil
}

func postgresesFromTeam(data gql.GetTeamPostgresesTeam, team string) []PostgresInstance {
	ret := make([]PostgresInstance, 0, len(data.Postgreses.Nodes))
	for _, pg := range data.Postgreses.Nodes {
		env := pg.TeamEnvironment.Environment.Name
		instance := PostgresInstance{
			Name:        output.Link{Name: pg.Name, URL: fmt.Sprintf("%s/team/%s/%s/postgres/%s", consoleBaseURL, team, env, pg.Name)},
			Environment: env, Version: pg.MajorVersion,
			HighAvailability: pg.HighAvailability,
		}
		if pg.ActiveBranch != nil {
			instance.ActiveBranch = pg.ActiveBranch.Name
		}
		ret = append(ret, instance)
	}
	return ret
}

func GetTeamPostgresBranches(ctx context.Context, team string, environments []string, labelFilters []gql.LabelFilter) ([]Instance, error) {
	_ = `# @genqlient
		query GetTeamPostgresBranches($team: Slug!, $postgresFilter: PostgresBranchFilter) {
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
	resp, err := gql.GetTeamPostgresBranches(ctx, client, team, &gql.PostgresBranchFilter{Environments: environments, Labels: labelFilters})
	if err != nil {
		return nil, err
	}
	return instancesFromTeam(resp.Team, team, environments), nil
}

func instancesFromTeam(data gql.GetTeamPostgresBranchesTeam, team string, environments []string) []Instance {
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
