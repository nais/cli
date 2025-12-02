package opensearch

import (
	"context"
	"strconv"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

type OpenSearch struct {
	// Memory is the memory for the OpenSearch instance.
	Memory gql.OpenSearchMemory `json:"memory" toml:"memory" jsonschema:"enum=GB_2,enum=GB_4,enum=GB_8,enum=GB_16,enum=GB_32,enum=GB_64"`
	// Tier is the tier of the OpenSearch instance.
	Tier gql.OpenSearchTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// Version is the major version of OpenSearch.
	Version gql.OpenSearchMajorVersion `json:"version,omitempty" toml:"version,omitempty" jsonschema:"enum=V2"`
	// StorageGB is the storage capacity in GB for the OpenSearch instance.
	StorageGB int `json:"storageGB,omitempty" toml:"storageGB,omitempty"`
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
		  $memory: OpenSearchMemory!,
		  $tier: OpenSearchTier!,
		  $version: OpenSearchMajorVersion!,
		  $storageGB: Int!,
		) {
		  createOpenSearch(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, memory: $memory, tier: $tier, version: $version, storageGB: $storageGB }
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

	resp, err := gql.CreateOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Memory, data.Tier, data.Version, data.StorageGB)
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
				memory
				tier
				storageGB
				version {
				  actual
				  desiredMajor
				}
				state
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
				memory
				tier
				storageGB
				version {
				  actual
				}
				state
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

func OpenSearchEnvironments(ctx context.Context, team, name string) ([]string, error) {
	all, err := GetAll(ctx, team)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0)
	for _, os := range all {
		if os.Name == name {
			ret = append(ret, os.TeamEnvironment.Environment.Name)
		}
	}
	return ret, nil
}

func Update(ctx context.Context, metadata Metadata, data *OpenSearch) (*gql.UpdateOpenSearchUpdateOpenSearchUpdateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient(omitempty: true)
		mutation UpdateOpenSearch(
		  $name: String!,
		  $environmentName: String!,
		  $teamSlug: Slug!,
		  $memory: OpenSearchMemory!,
		  $tier: OpenSearchTier!,
		  $version: OpenSearchMajorVersion!,
		  $storageGB: Int!,
		) {
		  updateOpenSearch(
		    input: { name: $name, environmentName: $environmentName, teamSlug: $teamSlug, memory: $memory, tier: $tier, version: $version, storageGB: $storageGB }
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

	resp, err := gql.UpdateOpenSearch(ctx, client, metadata.Name, metadata.EnvironmentName, metadata.TeamSlug, data.Memory, data.Tier, data.Version, data.StorageGB)
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
		{"Tier", string(openSearch.Tier)},
		{"Memory", string(openSearch.Memory)},
		{"Storage (GB)", strconv.Itoa(openSearch.StorageGB)},
		{"Version", openSearch.Version.Actual},
		{"State", string(openSearch.State)},
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
