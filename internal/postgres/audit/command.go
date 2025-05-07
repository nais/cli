package audit

import (
	"context"

	"github.com/nais/cli/internal/postgres"
)

func Run(ctx context.Context, applicationName string, flags postgres.Flags) error {
	return postgres.EnableAuditLogging(ctx, applicationName, flags.Context, flags.Namespace)
}
