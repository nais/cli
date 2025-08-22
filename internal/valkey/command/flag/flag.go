package flag

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
)

type Valkey struct {
	*alpha.Alpha
}

type Create struct {
	*Valkey
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
