package postgres

import (
	"github.com/nais/cli/internal/root"
)

type Flags struct {
	root.Flags
	Namespace string
	Context   string
}
