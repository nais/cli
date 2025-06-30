package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/metric"
	"github.com/nais/cli/v2/internal/postgres"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
)

func enableAuditCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.EnableAudit{Postgres: parentFlags}
	return &cli.Command{
		Name:        "enable-audit",
		Title:       "Enable audit extension in SQL instance database.",
		Description: "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		Args: []cli.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			err := postgres.EnableAuditLogging(ctx, args[0], flags.Context, flags.Namespace)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
			}
			return err
		},
	}
}
