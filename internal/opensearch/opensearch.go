package opensearch

import (
	"context"
	"strconv"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"k8s.io/utils/ptr"
)

type Metadata struct {
	// Name is the name of the OpenSearch instance.
	Name string
	// EnvironmentName is the name of the environment where the OpenSearch instance is created.
	EnvironmentName string
	// TeamSlug is the slug of the team that owns the OpenSearch instance.
	TeamSlug string
}

func Create(ctx context.Context, metadata Metadata, data gql.CreateOpenSearchInput) (*gql.CreateOpenSearchCreateOpenSearchCreateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient
		mutation CreateOpenSearch(
		  $input: CreateOpenSearchInput!
		) {
		  createOpenSearch(input: $input) {
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

	data.Name = metadata.Name
	data.EnvironmentName = metadata.EnvironmentName
	data.TeamSlug = metadata.TeamSlug
	resp, err := gql.CreateOpenSearch(ctx, client, data)
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

	return ptr.Deref(resp.DeleteOpenSearch.OpenSearchDeleted, false), nil
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
				shardIndexingPressureEnabled
				shardIndexingPressureEnforced
				indicesQueryBoolMaxClauseCount
				httpMaxContentLength
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

func GetAll(ctx context.Context, teamSlug string, filter gql.OpenSearchFilter) ([]gql.GetAllOpenSearchesTeamOpenSearchesOpenSearchConnectionNodesOpenSearch, error) {
	_ = `# @genqlient
		query GetAllOpenSearches($teamSlug: Slug!, $filter: OpenSearchFilter) {
		  team(slug: $teamSlug) {
			openSearches(filter: $filter) {
			  nodes {
				name
				memory
				tier
				storageGB
				shardIndexingPressureEnabled
				shardIndexingPressureEnforced
				indicesQueryBoolMaxClauseCount
				httpMaxContentLength
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

	resp, err := gql.GetAllOpenSearches(ctx, client, teamSlug, new(filter))
	if err != nil {
		return nil, err
	}
	return resp.Team.OpenSearches.Nodes, nil
}

func Update(ctx context.Context, metadata Metadata, data gql.UpdateOpenSearchInput) (*gql.UpdateOpenSearchUpdateOpenSearchUpdateOpenSearchPayloadOpenSearch, error) {
	_ = `# @genqlient
		mutation UpdateOpenSearch(
		  $input: UpdateOpenSearchInput!
		) {
		  updateOpenSearch(
		    input: $input
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

	data.Name = metadata.Name
	data.EnvironmentName = metadata.EnvironmentName
	data.TeamSlug = metadata.TeamSlug
	resp, err := gql.UpdateOpenSearch(ctx, client, data)
	if err != nil {
		return nil, err
	}

	return &resp.UpdateOpenSearch.OpenSearch, nil
}

func FormatDetails(metadata Metadata, openSearch *gql.GetOpenSearchTeamEnvironmentOpenSearch) [][]string {
	indicesQueryBoolMaxClauseCount := "(default)"
	if openSearch.IndicesQueryBoolMaxClauseCount != nil {
		indicesQueryBoolMaxClauseCount = strconv.Itoa(*openSearch.IndicesQueryBoolMaxClauseCount)
	}

	return [][]string{
		{"Field", "Value"},
		{"Team", metadata.TeamSlug},
		{"Environment", metadata.EnvironmentName},
		{"Name", metadata.Name},
		{"Tier", string(openSearch.Tier)},
		{"Memory", string(openSearch.Memory)},
		{"Storage (GB)", strconv.Itoa(openSearch.StorageGB)},
		{"Shard indexing pressure enabled", strconv.FormatBool(openSearch.ShardIndexingPressureEnabled)},
		{"Shard indexing pressure enforced", strconv.FormatBool(openSearch.ShardIndexingPressureEnforced)},
		{"Indices query bool max clause count", indicesQueryBoolMaxClauseCount},
		{"HTTP max content length", ptr.Deref(openSearch.HttpMaxContentLength, "(default)")},
		{"Version", ptr.Deref(openSearch.Version.Actual, "")},
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
			ptr.Deref(edge.Node.Workload.GetTypename(), ""),
			edge.Node.Access,
		})
	}
	return acl
}
