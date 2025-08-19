package apply

import "github.com/nais/cli/internal/naisapi/gql"

type Apply struct {
	Version          string `json:"naisVersion" toml:"naisVersion" jsonschema:"enum=v3"`
	ResourceMetadata `json:",inline"`

	// Valkey is a map of Valkey instances to be created, where the key is the name of the instance.
	Valkey map[string]*Valkey `json:"valkey,omitempty" toml:"valkey,omitempty"`
	// OpenSearch is a map of OpenSearch instances to be created, where the key is the name of the instance.
	OpenSearch map[string]*OpenSearch `json:"openSearch,omitempty" toml:"openSearch,omitempty"`
}

type ResourceMetadata struct {
	Environment string `json:"environment" toml:"environment"`
	TeamSlug    string `json:"team" toml:"team"`
}

type Valkey struct {
	// Size is the size of the Valkey instance.
	Size gql.ValkeySize `json:"size" toml:"size" jsonschema:"enum=RAM_1GB,enum=RAM_4GB,enum=RAM_8GB,enum=RAM_14GB,enum=RAM_28GB,enum=RAM_56GB,enum=RAM_112GB,enum=RAM_200GB"`
	// Tier is the tier of the Valkey instance.
	Tier gql.ValkeyTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// MaxMemoryPolicy is the max memory policy of the Valkey instance, e.g. "allkeys-lru".
	MaxMemoryPolicy gql.ValkeyMaxMemoryPolicy `json:"maxMemoryPolicy,omitempty" toml:"maxMemoryPolicy,omitempty"`
}

type OpenSearch struct {
	// Size is the size of the OpenSearch instance.
	Size gql.OpenSearchSize `json:"size" toml:"size" jsonschema:"enum=RAM_4GB,enum=RAM_8GB,enum=RAM_16GB,enum=RAM_32GB,enum=RAM_64GB"`
	// Tier is the tier of the OpenSearch instance.
	Tier gql.OpenSearchTier `json:"tier" toml:"tier" jsonschema:"enum=SINGLE_NODE,enum=HIGH_AVAILABILITY"`
	// Version is the major version of OpenSearch"
	Version gql.OpenSearchMajorVersion `json:"version,omitempty" toml:"version,omitempty" jsonschema:"enum=V2"`
}
