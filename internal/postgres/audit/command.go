package audit

import (
	"context"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.EnableAudit) error {
	err := postgres.EnableAuditLogging(ctx, applicationName, flags.Context, flags.Namespace)
	if err != nil {
		metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
	}
	return err
}
