package opensearch

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type OpenSearch struct {
	// Size is the size of the OpenSearch instance.
	Size gql.OpenSearchSize `json:"size" toml:"size" jsonschema:"enum=RAM_4GB,enum=RAM_8GB,enum=RAM_16GB,enum=RAM_32GB,enum=RAM_64GB"`
	// Tier is the tier of the OpenSearch instance.
	Tier gql.OpenSearchTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// Version is the major version of OpenSearch"
	Version gql.OpenSearchMajorVersion `json:"version,omitempty" toml:"version,omitempty" jsonschema:"enum=V2"`
}

func Upsert(ctx context.Context, name, environmentName, teamSlug string, data *OpenSearch) error {
	_, err := Create(ctx, name, environmentName, teamSlug, data)
	if naisapi.IsErrAlreadyExists(err) {
		_, err := Update(ctx, name, environmentName, teamSlug, data)
		return err
	}
	return err
}

func Create(ctx context.Context, name, environmentName, teamSlug string, data *OpenSearch) (*gql.CreateOpenSearchCreateOpenSearchCreateOpenSearchPayloadOpenSearch, error) {
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

	resp, err := gql.CreateOpenSearch(ctx, client, name, environmentName, teamSlug, data.Size, data.Tier, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.CreateOpenSearch.OpenSearch, nil
}

func Update(ctx context.Context, name, environmentName, teamSlug string, data *OpenSearch) (*gql.UpdateOpenSearchUpdateOpenSearchUpdateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient(omitempty: true)
		mutation UpdateOpenSearch(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $size: OpenSearchSize!,
		  $tier: OpenSearchTier!,
		  $version: OpenSearchMajorVersion,
		) {
		  updateOpenSearch(
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

	resp, err := gql.UpdateOpenSearch(ctx, client, name, environmentName, teamSlug, data.Size, data.Tier, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateOpenSearch.OpenSearch, nil
}
