package command

import (
	"context"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func enableAuditCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.EnableAudit{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "enable-audit",
		Title:       "Enable audit extension in SQL instance database.",
		Description: "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			err := postgres.EnableAuditLogging(ctx, args[0], flags.Context, flags.Namespace)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
			}
			return err
		},
	}
}
