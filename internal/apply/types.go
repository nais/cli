package apply

type Apply struct {
	Version     string       `json:"naisVersion" toml:"naisVersion" jsonschema:"enum=1"`
	Environment string       `json:"environment,omitempty" toml:"environment,omitempty"`
	Valkey      []Valkey     `json:"valkey,omitempty" toml:"valkey,omitempty"`
	OpenSearch  []OpenSearch `json:"openSearch,omitempty" toml:"openSearch,omitempty"`
}

type ValkeySize string

// allkeys-lfu, allkeys-lru, allkeys-random, noeviction, volatile-lfu, volatile-lru, volatile-random, volatile-ttl
type ValkeyMaxMemoryPolicy string

type Valkey struct {
	// Name is the name of the Valkey instance, e.g. "my-valkey".
	Name string `json:"name" toml:"name"`
	// Size is the size of the Valkey instance.
	Size ValkeySize `json:"size" toml:"size"`
	// MaxMemoryPolicy is the max memory policy of the Valkey instance, e.g. "allkeys-lru".
	MaxMemoryPolicy ValkeyMaxMemoryPolicy `json:"maxMemoryPolicy" toml:"maxMemoryPolicy"`
}

type OpenSearchSize string

type OpenSearch struct {
	Name string `json:"name" toml:"name"`
	// Size is the size of the OpenSearch instance.
	Size OpenSearchSize `json:"size" toml:"size"`
}
