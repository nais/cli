package psql

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.Psql) error {
	return postgres.RunPSQL(ctx, applicationName, flags.Context, flags.Namespace, flags.IsVerbose())
}
