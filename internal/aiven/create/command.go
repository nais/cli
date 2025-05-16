package create

import (
	"github.com/nais/cli/internal/aiven"
)

type Arguments struct {
	Username  string
	Namespace string
}

type Flags struct {
	*aiven.Flags
	Expire uint
	Secret string
}
