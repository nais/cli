package app

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetApplicationLabels(ctx context.Context, team, name, env string) ([]Label, error) {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	req := graphql.Request{
		OpName: "GetApplicationLabels",
		Query: `query GetApplicationLabels($team: Slug!, $name: String!, $env: [String!]) {
			team(slug: $team) {
				applications(filter: {name: $name, environments: $env}, first: 2) {
					nodes {
						...LabelledApplicationFields
					}
				}
			}
		}

		fragment LabelledApplicationFields on Application {
			name
			labels {
				key
				value
			}
		}`,
		Variables: struct {
			Team string   `json:"team"`
			Name string   `json:"name"`
			Env  []string `json:"env"`
		}{
			Team: team,
			Name: name,
			Env:  []string{env},
		},
	}

	respData := &struct {
		Team struct {
			Applications struct {
				Nodes []struct {
					Name   string  `json:"name"`
					Labels []Label `json:"labels"`
				} `json:"nodes"`
			} `json:"applications"`
		} `json:"team"`
	}{}
	resp := graphql.Response{Data: respData}
	if err := client.MakeRequest(ctx, &req, &resp); err != nil {
		return nil, err
	}

	nodes := respData.Team.Applications.Nodes
	if len(nodes) == 0 {
		return nil, fmt.Errorf("application %q not found in %q", name, env)
	}
	if len(nodes) > 1 {
		return nil, fmt.Errorf("expected one application, got %d", len(nodes))
	}

	labels := nodes[0].Labels
	slices.SortFunc(labels, func(a, b Label) int {
		return strings.Compare(a.Key, b.Key)
	})
	return labels, nil
}

func SetApplicationLabels(ctx context.Context, team, application, env string, updates map[string]string) ([]Label, error) {
	existing, err := GetApplicationLabels(ctx, team, application, env)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string, len(existing)+len(updates))
	for _, l := range existing {
		labels[l.Key] = l.Value
	}
	maps.Copy(labels, updates)

	labelInputs := make([]gql.ResourceLabelInput, 0, len(labels))
	keys := slices.Collect(maps.Keys(labels))
	slices.Sort(keys)
	for _, key := range keys {
		labelInputs = append(labelInputs, gql.ResourceLabelInput{
			Key:   key,
			Value: labels[key],
		})
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	req := graphql.Request{
		OpName: "SetApplicationLabels",
		Query: `mutation SetApplicationLabels($team: Slug!, $name: String!, $env: String!, $labels: [ResourceLabelInput!]!) {
			updateApplication(input: {teamSlug: $team, environmentName: $env, name: $name, labels: $labels}) {
				application {
					...LabelledApplicationFields
				}
			}
		}

		fragment LabelledApplicationFields on Application {
			name
			labels {
				key
				value
			}
		}`,
		Variables: struct {
			Team   string                   `json:"team"`
			Name   string                   `json:"name"`
			Env    string                   `json:"env"`
			Labels []gql.ResourceLabelInput `json:"labels"`
		}{
			Team:   team,
			Name:   application,
			Env:    env,
			Labels: labelInputs,
		},
	}

	respData := &struct {
		UpdateApplication struct {
			Application struct {
				Name   string  `json:"name"`
				Labels []Label `json:"labels"`
			} `json:"application"`
		} `json:"updateApplication"`
	}{}
	resp := graphql.Response{Data: respData}
	if err := client.MakeRequest(ctx, &req, &resp); err != nil {
		return nil, err
	}

	ret := respData.UpdateApplication.Application.Labels
	slices.SortFunc(ret, func(a, b Label) int {
		return strings.Compare(a.Key, b.Key)
	})
	return ret, nil
}
