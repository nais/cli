package valkey

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type Valkey struct {
	// Size is the size of the Valkey instance.
	Size gql.ValkeySize `json:"size" toml:"size" jsonschema:"enum=RAM_1GB,enum=RAM_4GB,enum=RAM_8GB,enum=RAM_14GB,enum=RAM_28GB,enum=RAM_56GB,enum=RAM_112GB,enum=RAM_200GB"`
	// Tier is the tier of the Valkey instance.
	Tier gql.ValkeyTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// MaxMemoryPolicy is the max memory policy of the Valkey instance, e.g. "allkeys-lru".
	MaxMemoryPolicy gql.ValkeyMaxMemoryPolicy `json:"maxMemoryPolicy,omitempty" toml:"maxMemoryPolicy,omitempty"`
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
		  $size: ValkeySize!,
		  $tier: ValkeyTier!,
		  $maxMemoryPolicy: ValkeyMaxMemoryPolicy,
		) {
		  createValkey(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, size: $size, tier: $tier, maxMemoryPolicy: $maxMemoryPolicy }
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

	resp, err := gql.CreateValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Size, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.CreateValkey.Valkey, nil
}

func Update(ctx context.Context, metadata Metadata, data *Valkey) (*gql.UpdateValkeyUpdateValkeyUpdateValkeyPayloadValkey, error) {
	_ = `# @genqlient(omitempty: true)
		mutation UpdateValkey(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $size: ValkeySize!,
		  $tier: ValkeyTier!,
		  $maxMemoryPolicy: ValkeyMaxMemoryPolicy,
		) {
		  updateValkey(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, size: $size, tier: $tier, maxMemoryPolicy: $maxMemoryPolicy }
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

	resp, err := gql.UpdateValkey(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Size, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateValkey.Valkey, nil
}
