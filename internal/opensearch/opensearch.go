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
	// Name is the name of the OpenSearch instance.
	Name string
	// EnvironmentName is the name of the environment where the OpenSearch instance is created.
	EnvironmentName string
	// TeamSlug is the slug of the team that owns the OpenSearch instance.
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

func Delete(ctx context.Context, metadata Metadata) (bool, error) {
	_ = `# @genqlient
		mutation DeleteOpenSearch($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  deleteOpenSearch(input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug }) {
		    openSearchDeleted
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return false, err
	}

	resp, err := gql.DeleteOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return false, err
	}

	return resp.DeleteOpenSearch.OpenSearchDeleted, nil
}

func Get(ctx context.Context, metadata Metadata) (*gql.GetOpenSearchTeamEnvironmentOpenSearch, error) {
	_ = `# @genqlient
		query GetOpenSearch($name: String!, $environmentName: String!, $teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			environment(name: $environmentName) {
			  openSearch(name: $name) {
				name
				size
				tier
				version
				majorVersion
				access(first: 1000, orderBy: {direction: ASC, field: ACCESS}) {
				  edges {
					node {
					  access
					  workload {
						id
						name
						__typename
						team {
						  slug
						}
					  }
					}
				  }
				}
			  }
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug)
	if err != nil {
		return nil, err
	}

	return &resp.Team.Environment.OpenSearch, nil
}

func GetAll(ctx context.Context, teamSlug string) ([]gql.GetAllOpenSearchesTeamOpenSearchesOpenSearchConnectionNodesOpenSearch, error) {
	_ = `# @genqlient
		query GetAllOpenSearches($teamSlug: Slug!) {
		  team(slug: $teamSlug) {
			openSearches {
			  nodes {
				name
				size
				tier
				version
				teamEnvironment {
				  environment {
					name
				  }
				}
				access(first: 1000) {
				  edges {
					node {
					  access
					}
				  }
				}
			  }
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.GetAllOpenSearches(ctx, client, teamSlug)
	if err != nil {
		return nil, err
	}
	return resp.Team.OpenSearches.Nodes, nil
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

func FormatDetails(metadata Metadata, openSearch *gql.GetOpenSearchTeamEnvironmentOpenSearch) [][]string {
	return [][]string{
		{"Field", "Value"},
		{"Team", metadata.TeamSlug},
		{"Environment", metadata.EnvironmentName},
		{"Name", metadata.Name},
		{"Size", string(openSearch.Size)},
		{"Tier", string(openSearch.Tier)},
		{"Version", openSearch.Version},
	}
}

func FormatAccessList(metadata Metadata, openSearch *gql.GetOpenSearchTeamEnvironmentOpenSearch) [][]string {
	acl := [][]string{
		{"Team", "Environment", "Name", "Type", "Access"},
	}
	for _, edge := range openSearch.Access.Edges {
		acl = append(acl, []string{
			edge.Node.Workload.GetTeam().Slug,
			metadata.EnvironmentName,
			edge.Node.Workload.GetName(),
			edge.Node.Workload.GetTypename(),
			edge.Node.Access,
		})
	}
	return acl
}
