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

type Metadata struct {
	// Name is the name of the Valkey instance.
	Name string
	// EnvironmentName is the name of the environment where the Valkey instance is created.
	EnvironmentName string
	// TeamSlug is the slug of the team that owns the Valkey instance.
	TeamSlug string
}

func Upsert(ctx context.Context, metadata Metadata, data *OpenSearch) error {
	_, err := Create(ctx, metadata, data)
	if naisapi.IsErrAlreadyExists(err) {
		_, err := Update(ctx, metadata, data)
		return err
	}
	return err
}

func Create(ctx context.Context, metadata Metadata, data *OpenSearch) (*gql.CreateOpenSearchCreateOpenSearchCreateOpenSearchPayloadOpenSearch, error) {
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

	resp, err := gql.CreateOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Size, data.Tier, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.CreateOpenSearch.OpenSearch, nil
}

func Update(ctx context.Context, metadata Metadata, data *OpenSearch) (*gql.UpdateOpenSearchUpdateOpenSearchUpdateOpenSearchPayloadOpenSearch, error) {
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

	resp, err := gql.UpdateOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Size, data.Tier, data.Version)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateOpenSearch.OpenSearch, nil
}
