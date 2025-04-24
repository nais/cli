package command

import (
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v2"
)

func usersList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List users in a Postgres database",
		ArgsUsage: "appname",
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

			return postgres.ListUsers(context.Context, appName, cluster, namespace)
		},
	}
}
