package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func audit() *cli.Command {
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 1 {
				metrics.AddOne(ctx, "postgres_missing_app_name_error_total")
				return ctx, fmt.Errorf("missing name of app")
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			appName := cmd.Args().First()

			namespace := cmd.String("namespace")
			cluster := cmd.String("context")

			return postgres.EnableAuditLogging(ctx, appName, cluster, namespace)
		},
	}
}
