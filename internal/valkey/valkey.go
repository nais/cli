package valkey

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Valkey struct {
	// Memory is the memory for the Valkey instance.
	Memory gql.ValkeyMemory `json:"memory" toml:"memory" jsonschema:"enum=GB_1,enum=GB_4,enum=GB_8,enum=GB_14,enum=GB_28,enum=GB_56,enum=GB_112,enum=GB_200"`
	// Tier is the tier of the Valkey instance.
	Tier gql.ValkeyTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// MaxMemoryPolicy is the max memory policy of the Valkey instance, e.g. "allkeys-lru".
	MaxMemoryPolicy gql.ValkeyMaxMemoryPolicy `json:"maxMemoryPolicy,omitempty" toml:"maxMemoryPolicy,omitempty" jsonschema:"enum=ALLKEYS_LFU,enum=ALLKEYS_LRU,enum=ALLKEYS_RANDOM,enum=NO_EVICTION,enum=VOLATILE_LFU,enum=VOLATILE_LRU,enum=VOLATILE_RANDOM,enum=VOLATILE_TTL"`
}

type Metadata struct {
	// Name is the name of the Valkey instance.
	Name string
	// EnvironmentName is the name of the environment where the Valkey instance is created.
	EnvironmentName string
	// TeamSlug is the slug of the team that owns the Valkey instance.
	TeamSlug string
}

func Upsert(ctx context.Context, metadata Metadata, data *Valkey) error {
	_, err := Create(ctx, metadata, data)
	if naisapi.IsErrAlreadyExists(err) {
		_, err := Update(ctx, metadata, data)
		return err
	}
	return err
}

func Create(ctx context.Context, metadata Metadata, data *Valkey) (*gql.CreateValkeyCreateValkeyCreateValkeyPayloadValkey, error) {
	_ = `# @genqlient(omitempty: true)
		mutation CreateValkey(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $memory: ValkeyMemory!,
		  $tier: ValkeyTier!,
		  $maxMemoryPolicy: ValkeyMaxMemoryPolicy,
		) {
		  createValkey(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, memory: $memory, tier: $tier, maxMemoryPolicy: $maxMemoryPolicy }
		  ) {
		    valkey {
		      id
		      name
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.CreateValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Memory, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.CreateValkey.Valkey, nil
}

func Delete(ctx context.Context, metadata Metadata) (bool, error) {
	_ = `# @genqlient
		mutation DeleteValkey($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  deleteValkey(input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug }) {
		    valkeyDeleted
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}

	resp, err := gql.DeleteValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return false, err
	}

	return resp.DeleteValkey.ValkeyDeleted, nil
}

func Get(ctx context.Context, metadata Metadata) (*gql.GetValkeyTeamEnvironmentValkey, error) {
	_ = `# @genqlient
		query GetValkey($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			environment(name: $environmentName) {
			  valkey(name: $name) {
				name
				memory
				tier
				maxMemoryPolicy
				state
				access(first: 1000, orderBy: {direction: ASC, field: ACCESS}) {
				  edges {
					node {
					  access
					  workload {
						id
						name
						__typename
						team {
						  slug
						}
					  }
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

	resp, err := gql.GetValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.Team.Environment.Valkey, nil
}

func GetAll(ctx context.Context, teamSlug string) ([]gql.GetAllValkeysTeamValkeysValkeyConnectionNodesValkey, error) {
	_ = `# @genqlient
		query GetAllValkeys($teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			valkeys {
			  nodes {
				name
				memory
				tier
				maxMemoryPolicy
				state
				teamEnvironment {
				  environment {
					name
				  }
				}
				access(first: 1000) {
				  edges {
					node {
					  access
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

	resp, err := gql.GetAllValkeys(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}

	return resp.Team.Valkeys.Nodes, nil
}

func Update(ctx context.Context, metadata Metadata, data *Valkey) (*gql.UpdateValkeyUpdateValkeyUpdateValkeyPayloadValkey, error) {
	_ = `# @genqlient(omitempty: true)
		mutation UpdateValkey(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $memory: ValkeyMemory!,
		  $tier: ValkeyTier!,
		  $maxMemoryPolicy: ValkeyMaxMemoryPolicy,
		) {
		  updateValkey(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, memory: $memory, tier: $tier, maxMemoryPolicy: $maxMemoryPolicy }
		  ) {
		    valkey {
		      id
		      name
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.UpdateValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Memory, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateValkey.Valkey, nil
}

func FormatDetails(metadata Metadata, valkey *gql.GetValkeyTeamEnvironmentValkey) [][]string {
	return [][]string{
		{"Field", "Value"},
		{"Team", metadata.TeamSlug},
		{"Environment", metadata.EnvironmentName},
		{"Name", metadata.Name},
		{"Memory", string(valkey.Memory)},
		{"Tier", string(valkey.Tier)},
		{"Max memory policy", string(valkey.MaxMemoryPolicy)},
		{"State", string(valkey.State)},
	}
}

func FormatAccessList(metadata Metadata, valkey *gql.GetValkeyTeamEnvironmentValkey) [][]string {
	acl := [][]string{
		{"Team", "Environment", "Name", "Type", "Access"},
	}
	for _, edge := range valkey.Access.Edges {
		acl = append(acl, []string{
			edge.Node.Workload.GetTeam().Slug,
			metadata.EnvironmentName,
			edge.Node.Workload.GetName(),
			edge.Node.Workload.GetTypename(),
			edge.Node.Access,
		})
	}
	return acl
}
