package postgrescmd

import (
	"fmt"
	"github.com/nais/cli/pkg/metrics"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func usersListCommand() *cli.Command {
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
			metrics.AddOne("postgres", "postgres_users_list_total")
			if context.Args().Len() < 1 {
				metrics.AddOne("nais_cli", "postgres_missing_app_name_error_total")
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
