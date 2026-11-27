package workload

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// SearchNames returns names for shell completion, not a list of valid workload names.
func SearchNames(ctx context.Context, kind gql.SearchType, team, environment, prefix string) ([]string, error) {
	_ = `# @genqlient
		query SearchWorkloadNames($filter: SearchFilter!) {
			search(first: 30, filter: $filter) {
				nodes {
					... on Application { name teamEnvironment { environment { name } } }
					... on Job { name teamEnvironment { environment { name } } }
				}
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := gql.SearchWorkloadNames(ctx, client, gql.SearchFilter{
		Query: prefix, Types: []gql.SearchType{kind}, TeamScope: &gql.TeamScopeInput{TeamSlugs: []string{team}},
	})
	if err != nil {
		return nil, err
	}

	// SearchFilter scopes by team and kind, but not by environment. Search can also
	// return fuzzy matches; only suggest names the shell can complete from the prefix.
	var names []string
	seen := make(map[string]bool)
	for _, node := range resp.Search.Nodes {
		var name, env string
		switch n := node.(type) {
		case *gql.SearchWorkloadNamesSearchSearchNodeConnectionNodesApplication:
			name, env = n.Name, n.TeamEnvironment.Environment.Name
		case *gql.SearchWorkloadNamesSearchSearchNodeConnectionNodesJob:
			name, env = n.Name, n.TeamEnvironment.Environment.Name
		default:
			continue
		}
		if (environment == "" || env == environment) && strings.HasPrefix(name, prefix) && !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names, nil
}
