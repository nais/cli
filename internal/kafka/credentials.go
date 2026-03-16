package kafka

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func CreateCredentials(ctx context.Context, teamSlug, environmentName, ttl string) (*gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials, error) {
	_ = `# @genqlient
		mutation CreateKafkaCredentials(
		  $teamSlug: Slug!,
		  $environmentName: String!,
		  $ttl: String!,
		) {
		  createKafkaCredentials(
		    input: { teamSlug: $teamSlug, environmentName: $environmentName, ttl: $ttl }
		  ) {
		    credentials {
		      username
		      accessCert
		      accessKey
		      caCert
		      brokers
		      schemaRegistry
		    }
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.CreateKafkaCredentials(ctx, client, teamSlug, environmentName, ttl)
	if err != nil {
		return nil, err
	}

	return &resp.CreateKafkaCredentials.Credentials, nil
}
