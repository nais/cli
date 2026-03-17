package opensearch

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func CreateCredentials(ctx context.Context, teamSlug, environmentName, instanceName string, permission gql.CredentialPermission, ttl string) (*gql.CreateOpenSearchCredentialsCreateOpenSearchCredentialsCreateOpenSearchCredentialsPayloadCredentialsOpenSearchCredentials, error) {
	_ = `# @genqlient
		mutation CreateOpenSearchCredentials(
		  $teamSlug: Slug!,
		  $environmentName: String!,
		  $instanceName: String!,
		  $permission: CredentialPermission!,
		  $ttl: String!,
		) {
		  createOpenSearchCredentials(
		    input: { teamSlug: $teamSlug, environmentName: $environmentName, instanceName: $instanceName, permission: $permission, ttl: $ttl }
		  ) {
		    credentials {
		      username
		      password
		      host
		      port
		      uri
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.CreateOpenSearchCredentials(ctx, client, teamSlug, environmentName, instanceName, permission, ttl)
	if err != nil {
		return nil, err
	}

	return &resp.CreateOpenSearchCredentials.Credentials, nil
}
