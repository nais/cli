package command

import (
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v2"
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
			verbose := context.Bool("verbose")

			return postgres.RunPSQL(context.Context, appName, cluster, namespace, verbose)
		},
	}
}
