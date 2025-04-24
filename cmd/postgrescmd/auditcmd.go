package postgrescmd

import (
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v2"
)

func auditCommand() *cli.Command {
	return &cli.Command{
		Name:        "enable-audit",
		Usage:       "Enable audit extension in Postgres database",
		Description: "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		ArgsUsage:   "appname",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 1 {
				metrics.AddOne("postgres_missing_app_name_error_total")
				return fmt.Errorf("missing name of app")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().First()

			namespace := context.String("namespace")
			cluster := context.String("context")

			return postgres.EnableAuditLogging(context.Context, appName, cluster, namespace)
		},
	}
}
