package resource

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"k8s.io/apimachinery/pkg/api/resource"
)

func init() {
	register(openSearchResource{kindSupport{kind: "OpenSearch", strippedVersion: "v1"}})
}

// openSearchResource applies OpenSearch instances through the nais-api
// OpenSearch mutation. It is only available as a stripped manifest.
type openSearchResource struct{ kindSupport }

// openSearchSpec is the user-facing (CRD-flavoured) OpenSearch spec.
type openSearchSpec struct {
	Memory                string                           `yaml:"memory"`
	Tier                  string                           `yaml:"tier"`
	Version               string                           `yaml:"version"`
	StorageGB             int                              `yaml:"storageGB"`
	ShardIndexingPressure *OpenSearchShardIndexingPressure `yaml:"shardIndexingPressure"`
	Indices               *OpenSearchIndices               `yaml:"indices"`
	Http                  *OpenSearchHttp                  `yaml:"http"`
}

type OpenSearchShardIndexingPressure struct {
	Enabled  bool `yaml:"enabled"`
	Enforced bool `yaml:"enforced"`
}

type OpenSearchIndices struct {
	QueryBoolMaxClauseCount *int `yaml:"queryBoolMaxClauseCount"`
}

type OpenSearchHttp struct {
	MaxContentLength *resource.Quantity `yaml:"maxContentLength"`
}

var (
	openSearchMemory = map[string]gql.OpenSearchMemory{
		"2GB":  gql.OpenSearchMemoryGb2,
		"4GB":  gql.OpenSearchMemoryGb4,
		"8GB":  gql.OpenSearchMemoryGb8,
		"16GB": gql.OpenSearchMemoryGb16,
		"32GB": gql.OpenSearchMemoryGb32,
		"64GB": gql.OpenSearchMemoryGb64,
	}
	openSearchTier = map[string]gql.OpenSearchTier{
		"SingleNode":       gql.OpenSearchTierSingleNode,
		"HighAvailability": gql.OpenSearchTierHighAvailability,
	}
	openSearchVersion = map[string]gql.OpenSearchMajorVersion{
		"1":    gql.OpenSearchMajorVersionV1,
		"2":    gql.OpenSearchMajorVersionV2,
		"2.19": gql.OpenSearchMajorVersionV219,
		"3.3":  gql.OpenSearchMajorVersionV33,
		"3.6":  gql.OpenSearchMajorVersionV36,
	}
)

func (o openSearchResource) Apply(ctx context.Context, meta Metadata, m Manifest) (Action, error) {
	var s openSearchSpec
	if err := decodeSpec(&m.Spec, &s); err != nil {
		return "", err
	}

	data := gql.CreateOpenSearchInput{
		Name:            m.Name,
		EnvironmentName: meta.EnvironmentName,
		TeamSlug:        meta.TeamSlug,
		StorageGB:       s.StorageGB,
	}

	var err error
	if data.Memory, err = enumValue("memory", s.Memory, openSearchMemory); err != nil {
		return "", err
	}
	if data.Tier, err = enumValue("tier", s.Tier, openSearchTier); err != nil {
		return "", err
	}
	if data.Version, err = enumValue("version", s.Version, openSearchVersion); err != nil {
		return "", err
	}
	if s.ShardIndexingPressure != nil {
		data.ShardIndexingPressureEnabled = new(s.ShardIndexingPressure.Enabled)
		data.ShardIndexingPressureEnforced = new(s.ShardIndexingPressure.Enforced)
	}
	if s.Indices != nil {
		data.IndicesQueryBoolMaxClauseCount = s.Indices.QueryBoolMaxClauseCount
	}
	if s.Http != nil && s.Http.MaxContentLength != nil {
		maxContentLength := s.Http.MaxContentLength.String()
		data.HttpMaxContentLength = &maxContentLength
	}

	ometa := opensearch.Metadata{
		Name:            m.Name,
		EnvironmentName: meta.EnvironmentName,
		TeamSlug:        meta.TeamSlug,
	}

	exists, err := o.exists(ctx, ometa)
	if err != nil {
		return "", err
	}

	if exists {
		if _, err := opensearch.Update(ctx, ometa, gql.UpdateOpenSearchInput{
			Name:                           data.Name,
			EnvironmentName:                data.EnvironmentName,
			TeamSlug:                       data.TeamSlug,
			Tier:                           data.Tier,
			Memory:                         data.Memory,
			Version:                        data.Version,
			StorageGB:                      data.StorageGB,
			ShardIndexingPressureEnabled:   data.ShardIndexingPressureEnabled,
			ShardIndexingPressureEnforced:  data.ShardIndexingPressureEnforced,
			IndicesQueryBoolMaxClauseCount: data.IndicesQueryBoolMaxClauseCount,
			HttpMaxContentLength:           data.HttpMaxContentLength,
		}); err != nil {
			return "", err
		}
		return ActionUpdated, nil
	}
	if _, err := opensearch.Create(ctx, ometa, data); err != nil {
		return "", err
	}
	return ActionCreated, nil
}

// exists reports whether an OpenSearch instance with the given name already exists.
func (o openSearchResource) exists(ctx context.Context, meta opensearch.Metadata) (bool, error) {
	_, err := opensearch.Get(ctx, meta)
	if err != nil {
		if naisapi.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
