package flag

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
)

type Valkey struct {
	*alpha.Alpha
}

type Create struct {
	*Valkey
	Size            string `name:"size" short:"s" usage:"Size of the Valkey instance."`
	Tier            string `name:"tier" short:"t" usage:"Tier of the Valkey instance."`
	MaxMemoryPolicy string `name:"max-memory-policy" short:"m" usage:"Max memory policy for the Valkey instance."`
}

type Delete struct {
	*Valkey
}

type Describe struct {
	*Valkey
}

type List struct {
	*Valkey
}

type Update struct {
	*Valkey
}
