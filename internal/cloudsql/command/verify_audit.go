package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func verifyAuditCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.VerifyAudit{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:        "verify-audit",
		Title:       "Verify audit extension and configuration in SQL instance database.",
		Description: "This verifies that the pgaudit extension is installed and that audit logging is properly configured for the application user.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			err := cloudsql.VerifyAuditLogging(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "verify_audit_logging_error")
			}
			return err
		},
	}
}
