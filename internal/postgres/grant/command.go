package grant

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.Grant) error {
	return postgres.GrantAndCreateSQLUser(ctx, applicationName, flags.Context, flags.Namespace)
}
