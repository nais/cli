package app

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type ValueSource struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func (s ValueSource) String() string {
	switch s.Kind {
	case "SECRET":
		return fmt.Sprintf("Secret/%s", s.Name)
	case "CONFIG":
		return fmt.Sprintf("Config/%s", s.Name)
	case "NAIS":
		return fmt.Sprintf("Nais/%s", s.Name)
	case "SPEC":
		return s.Name
	default:
		return fmt.Sprintf("%s/%s", s.Kind, s.Name)
	}
}

type EnvVarValue struct {
	Value    string
	IsSecret bool
}

func (v EnvVarValue) String() string {
	if v.IsSecret {
		return "●●●●●●●●"
	}
	return v.Value
}

func (v EnvVarValue) MarshalJSON() ([]byte, error) {
	if v.IsSecret {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "%q", v.Value), nil
}

type EnvVar struct {
	Name   string      `json:"name"`
	Value  EnvVarValue `json:"value"`
	Source ValueSource `json:"source"`
}

func GetApplicationEnvVars(ctx context.Context, slug, name string, envs []string) ([]EnvVar, error) {
	_ = `# @genqlient
		query GetApplicationEnvVars($slug: Slug!, $name: String!, $env: [String!]) {
		  team(slug: $slug) {
		    applications(filter: { name: $name, environments: $env }) {
		      nodes {
		        instanceGroups {
		          created
		          environmentVariables {
		            name
		            value
		            source {
		              kind
		              name
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

	resp, err := gql.GetApplicationEnvVars(ctx, client, slug, name, envs)
	if err != nil {
		return nil, err
	}

	if len(resp.Team.Applications.Nodes) == 0 {
		return nil, fmt.Errorf("application %q not found", name)
	}

	groups := resp.Team.Applications.Nodes[0].InstanceGroups
	if len(groups) == 0 {
		return nil, nil
	}

	newest := groups[0]
	for _, g := range groups[1:] {
		if g.Created.After(newest.Created) {
			newest = g
		}
	}

	ret := make([]EnvVar, 0, len(newest.EnvironmentVariables))
	for _, ev := range newest.EnvironmentVariables {
		isSecret := ev.Source.Kind == gql.InstanceGroupValueSourceKindSecret
		ret = append(ret, EnvVar{
			Name: ev.Name,
			Value: EnvVarValue{
				Value:    ev.Value,
				IsSecret: isSecret,
			},
			Source: ValueSource{
				Kind: string(ev.Source.Kind),
				Name: ev.Source.Name,
			},
		})
	}
	return ret, nil
}

// UniqueSecretNames returns deduplicated secret names from env vars.
func UniqueSecretNames(vars []EnvVar) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, v := range vars {
		if v.Source.Kind == "SECRET" {
			if _, ok := seen[v.Source.Name]; !ok {
				seen[v.Source.Name] = struct{}{}
				names = append(names, v.Source.Name)
			}
		}
	}
	return names
}
