package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/audit"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func enableAuditCommand(parentFlags *flag.Postgres) *cli.Command {
	return cli.NewCommand("enable-audit", "Enable audit extension in SQL instance database.",
		cli.WithLong("This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts."),
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return audit.Run(ctx, args[0], &flag.EnableAudit{Postgres: parentFlags})
		}),
	)
}
