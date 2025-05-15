package audit

import (
	"context"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/postgres"
)

func Run(ctx context.Context, applicationName string, flags postgres.Flags) error {
	err := postgres.EnableAuditLogging(ctx, applicationName, flags.Context, flags.Namespace)
	if err != nil {
		metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
	}
	return err
}
