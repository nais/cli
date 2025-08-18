package apply

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func CreateOpenSearch(ctx context.Context, name string, metadata ResourceMetadata, data *OpenSearch) (*gql.CreateOpenSearchCreateOpenSearchCreateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient(omitempty: true)
		mutation CreateOpenSearch(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $size: OpenSearchSize!,
		  $version: OpenSearchMajorVersion,
		) {
		  createOpenSearch(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, size: $size, version: $version }
		  ) {
		    openSearch {
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

	resp, err := gql.CreateOpenSearch(ctx, client, name, metadata.Environment, metadata.TeamSlug, data.Size, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.CreateOpenSearch.OpenSearch, nil
}
