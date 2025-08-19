package apply

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func UpsertValkey(ctx context.Context, name string, metadata ResourceMetadata, data *Valkey) error {
	_, err := CreateValkey(ctx, name, metadata, data)
	if err != nil {
		if naisapi.IsAlreadyExistsError(err) {
			return UpdateValkey(ctx, name, metadata, data)
		}
		return err
	}
	return nil
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

func UpdateValkey(ctx context.Context, name string, metadata ResourceMetadata, data *Valkey) error {
	return fmt.Errorf("update valkey is not implemented")
}
