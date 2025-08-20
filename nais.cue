package nais

// cue import nais.toml -o nais.cue -p nais

naisVersion: "v3"
environment: "dev"
team:        "devteam"

valkey: "24s": {
	size:            "RAM_14GB"
	tier:            "SINGLE_NODE"
	maxMemoryPolicy: "ALLKEYS_LRU"
}

openSearch: asdf: {
	size: "RAM_4GB"
	tier: "SINGLE_NODE"
}

openSearch: foo: {
	size: "RAM_32GB"
	tier: "SINGLE_NODE"
}

openSearch: bar: {
	size: "RAM_6a4GB"
	tier: "HIGH_AVAILABILITY"
}
