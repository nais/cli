package psql

import (
	"context"

	"github.com/nais/cli/internal/postgres"
)

func Run(ctx context.Context, applicationName string, flags *postgres.Flags) error {
	return postgres.RunPSQL(ctx, applicationName, flags.Context, flags.Namespace, flags.IsVerbose())
}
