package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func enableAuditCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.EnableAudit{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:        "enable-audit",
		Title:       "Enable audit extension in SQL instance database.",
		Description: "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database should have auditing enabled"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			err := cloudsql.EnableAuditLogging(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "enable_audit_logging_error")
			}
			return err
		},
	}
}
