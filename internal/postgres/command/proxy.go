package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func proxy() *cli.Command {
	return &cli.Command{
		Name:        "proxy",
		Usage:       "Create a proxy to a Postgres instance",
		Description: "Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.",
		ArgsUsage:   "appname",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   5432,
			},
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   "localhost",
			},
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
			port := cmd.Uint("port")
			host := cmd.String("host")

			return postgres.RunProxy(ctx, appName, cluster, namespace, host, port, verbose)
		},
	}
}
