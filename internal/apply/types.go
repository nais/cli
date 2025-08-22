package apply

import (
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
)

type Apply struct {
	Version     string `json:"naisVersion" toml:"naisVersion" jsonschema:"enum=v3"`
	Environment string `json:"environment" toml:"environment"`
	TeamSlug    string `json:"team" toml:"team"`

	// Valkey is a map of Valkey instances to be created, where the key is the name of the instance.
	Valkey map[string]*valkey.Valkey `json:"valkey,omitempty" toml:"valkey,omitempty"`
	// OpenSearch is a map of OpenSearch instances to be created, where the key is the name of the instance.
	OpenSearch map[string]*OpenSearch `json:"openSearch,omitempty" toml:"openSearch,omitempty"`
}

type OpenSearch struct {
	// Size is the size of the OpenSearch instance.
	Size gql.OpenSearchSize `json:"size" toml:"size" jsonschema:"enum=RAM_4GB,enum=RAM_8GB,enum=RAM_16GB,enum=RAM_32GB,enum=RAM_64GB"`
	// Tier is the tier of the OpenSearch instance.
	Tier gql.OpenSearchTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// Version is the major version of OpenSearch"
	Version gql.OpenSearchMajorVersion `json:"version,omitempty" toml:"version,omitempty" jsonschema:"enum=V2"`
}
