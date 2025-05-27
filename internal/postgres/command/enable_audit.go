package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func enableAuditCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.EnableAudit{Postgres: parentFlags}
	return cli.NewCommand("enable-audit", "Enable audit extension in SQL instance database.",
		cli.WithLong("This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts."),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			err := postgres.EnableAuditLogging(ctx, args[0], flags.Context, flags.Namespace)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
			}
			return err
		}),
	)
}
