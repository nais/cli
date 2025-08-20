package apply

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func UpsertValkey(ctx context.Context, name string, metadata ResourceMetadata, data *Valkey) error {
	_, err := CreateValkey(ctx, name, metadata, data)
	if naisapi.IsErrAlreadyExists(err) {
		_, err := UpdateValkey(ctx, name, metadata, data)
		return err
	}
	return err
}

func CreateValkey(ctx context.Context, name string, metadata ResourceMetadata, data *Valkey) (*gql.CreateValkeyCreateValkeyCreateValkeyPayloadValkey, error) {
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

	resp, err := gql.CreateValkey(ctx, client, name, metadata.Environment, metadata.TeamSlug, data.Size, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.CreateValkey.Valkey, nil
}

func UpdateValkey(ctx context.Context, name string, metadata ResourceMetadata, data *Valkey) (*gql.UpdateValkeyUpdateValkeyUpdateValkeyPayloadValkey, error) {
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

	resp, err := gql.UpdateValkey(ctx, client, name, metadata.Environment, metadata.TeamSlug, data.Size, data.Tier, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateValkey.Valkey, nil
}
