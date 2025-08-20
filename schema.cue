package nais

// cue import schema.json -o schema.cue -p nais

@jsonschema(schema="https://json-schema.org/draft/2020-12/schema")
@jsonschema(id="https://nais.io/schema/apply")
close({
	environment!: string
	naisVersion!: "v3"

	// OpenSearch is a map of OpenSearch instances to be created,
	// where the key is the name of the instance.
	openSearch?: [string]: close({
		// Size is the size of the OpenSearch instance.
		size!: "RAM_4GB" | "RAM_8GB" | "RAM_16GB" | "RAM_32GB" | "RAM_64GB"

		// Tier is the tier of the OpenSearch instance.
		tier!: "SINGLE_NODE" | "HIGH_AVAILABILITY"

		// Version is the major version of OpenSearch"
		version?: "V2"
	})
	team!: string

	// Valkey is a map of Valkey instances to be created, where the
	// key is the name of the instance.
	valkey?: [string]: close({
		// MaxMemoryPolicy is the max memory policy of the Valkey
		// instance, e.g. "allkeys-lru".
		maxMemoryPolicy?: string

		// Size is the size of the Valkey instance.
		size!: "RAM_1GB" | "RAM_4GB" | "RAM_8GB" | "RAM_14GB" | "RAM_28GB" | "RAM_56GB" | "RAM_112GB" | "RAM_200GB"

		// Tier is the tier of the Valkey instance.
		tier!: "SINGLE_NODE" | "HIGH_AVAILABILITY"
	})
})
