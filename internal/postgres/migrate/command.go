package migrate

import (
	"github.com/nais/cli/internal/postgres"
)

type Arguments struct {
	ApplicationName    string
	TargetInstanceName string
}

type Flags struct {
	*postgres.Flags
	DryRun bool
}
