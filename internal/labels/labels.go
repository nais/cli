package labels

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

// teamProvider is a minimal interface for flag structs that carry a team value.
// All flag structs that embed *flags.AdditionalFlags satisfy this interface.
type teamProvider interface {
	GetTeam() string
}

// LabelFacetResource is implemented by flag structs to declare which resource
// type's label facets should be queried for tab completion.
type LabelFacetResource interface {
	LabelFacetResource() string
}

// LabelFilters is a []string flag type with autocomplete support.
// It completes KEY=VALUE pairs drawn from the team's existing resource labels.
// The flags struct must implement [LabelFacetResource] to select which resource to query.
type LabelFilters []string

var _ naistrix.FlagAutoCompleter = (*LabelFilters)(nil)

func (l *LabelFilters) AutoComplete(ctx context.Context, _ *naistrix.Arguments, toComplete string, flags any) ([]string, string) {
	tp, ok := flags.(teamProvider)
	if !ok {
		return nil, ""
	}
	team := tp.GetTeam()
	if team == "" {
		return nil, "Provide --team to get label suggestions."
	}

	rp, ok := flags.(LabelFacetResource)
	if !ok {
		return nil, ""
	}

	pairs, err := getTeamLabelFacets(ctx, team, rp.LabelFacetResource())
	if err != nil {
		return nil, fmt.Sprintf("Unable to fetch label suggestions: %v", err)
	}

	var completions []string
	for _, kv := range pairs {
		if strings.HasPrefix(kv, toComplete) {
			completions = append(completions, kv)
		}
	}
	return completions, "Available labels in KEY=VALUE form."
}

func getTeamLabelFacets(ctx context.Context, team, resource string) ([]string, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	// Build a query that fetches only the requested resource's label facets.
	// The resource name must be a valid field on the Team type in the schema.
	query := fmt.Sprintf(`query GetTeamLabelFacets($team: Slug!) {
		team(slug: $team) {
			%s { facets { labels { key value } } }
		}
	}`, resource)

	type labelItem struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	type facets struct {
		Labels []labelItem `json:"labels"`
	}
	type conn struct {
		Facets facets `json:"facets"`
	}
	respData := &struct {
		Team map[string]conn `json:"team"`
	}{}
	resp := graphql.Response{Data: respData}
	if err := client.MakeRequest(ctx, &graphql.Request{
		OpName: "GetTeamLabelFacets",
		Query:  query,
		Variables: struct {
			Team string `json:"team"`
		}{Team: team},
	}, &resp); err != nil {
		return nil, err
	}

	c := respData.Team[resource]
	pairs := make([]string, 0, len(c.Facets.Labels))
	for _, l := range c.Facets.Labels {
		pairs = append(pairs, l.Key+"="+l.Value)
	}
	slices.Sort(pairs)
	return pairs, nil
}

func ParseAssignments(args LabelFilters) (map[string]string, error) {
	labels := make(map[string]string, len(args))
	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid label argument: %q (expected KEY=VALUE)", arg)
		}
		labels[key] = value
	}
	return labels, nil
}

func ParseFilters(args LabelFilters) ([]gql.LabelFilter, error) {
	assignments, err := ParseAssignments(args)
	if err != nil {
		return nil, err
	}

	keys := slices.Collect(maps.Keys(assignments))
	slices.Sort(keys)

	filters := make([]gql.LabelFilter, 0, len(keys))
	for _, key := range keys {
		filters = append(filters, gql.LabelFilter{
			Key:   key,
			Value: assignments[key],
		})
	}
	return filters, nil
}
