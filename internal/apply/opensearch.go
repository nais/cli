package apply

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

func UpsertOpenSearch(ctx context.Context, name string, metadata ResourceMetadata, data *OpenSearch) error {
	_, err := CreateOpenSearch(ctx, name, metadata, data)
	if err != nil {
		if naisapi.IsAlreadyExistsError(err) {
			return UpdateOpenSearch(ctx, name, metadata, data)
		}
		return err
	}
	return nil
}

func CreateOpenSearch(ctx context.Context, name string, metadata ResourceMetadata, data *OpenSearch) (*gql.CreateOpenSearchCreateOpenSearchCreateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient(omitempty: true)
		mutation CreateOpenSearch(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $size: OpenSearchSize!,
		  $tier: OpenSearchTier!,
		  $version: OpenSearchMajorVersion,
		) {
		  createOpenSearch(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, size: $size, tier: $tier, version: $version }
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

	resp, err := gql.CreateOpenSearch(ctx, client, name, metadata.Environment, metadata.TeamSlug, data.Size, data.Tier, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.CreateOpenSearch.OpenSearch, nil
}

func UpdateOpenSearch(ctx context.Context, name string, metadata ResourceMetadata, data *OpenSearch) error {
	return fmt.Errorf("update opensearch is not implemented")
}
