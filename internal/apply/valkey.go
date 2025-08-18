package apply

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func CreateValkey(ctx context.Context, data *Valkey) (*gql.CreateValkeyCreateValkeyCreateValkeyPayloadValkey, error) {
	_ = `# @genqlient
		mutation CreateValkey(
		  $name: String!,
		  $size: ValkeySize!,
		  $maxMemoryPolicy: ValkeyMaxMemoryPolicy,
		) {
		  createValkey(
		    input: { name: $name, size: $size, maxMemoryPolicy: $maxMemoryPolicy }
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

	resp, err := gql.CreateValkey(ctx, client, data.Name, data.Size, data.MaxMemoryPolicy)
	if err != nil {
		return nil, err
	}

	return &resp.CreateValkey.Valkey, nil
}
