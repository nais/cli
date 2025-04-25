package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func psql() *cli.Command {
	return &cli.Command{
		Name:        "psql",
		Usage:       "Connect to the database using psql",
		Description: "Create a shell to the postgres instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		ArgsUsage:   "appname",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
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
			verbose := cmd.Bool("verbose")

			return postgres.RunPSQL(ctx, appName, cluster, namespace, verbose)
		},
	}
}
