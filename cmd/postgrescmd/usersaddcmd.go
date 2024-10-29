package postgrescmd

import (
	"fmt"
	"github.com/nais/cli/pkg/metrics"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func usersAddCommand() *cli.Command {
	return &cli.Command{
		Name:        "add",
		Usage:       "Add user to a Postgres database",
		Description: "Will grant a user access to tables in public schema.",
		ArgsUsage:   "appname username password",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "privilege",
				Aliases: []string{"p"},
				Value:   "select",
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
			metrics.AddOne("postgres", "postgres_users_add_total")
			if context.Args().Len() < 3 {
				return fmt.Errorf("missing required arguments: appname, username, password")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().Get(0)
			username := context.Args().Get(1)
			password := context.Args().Get(2)

			namespace := context.String("namespace")
			cluster := context.String("context")
			privilege := context.String("privilege")

			return postgres.AddUser(context.Context, appName, username, password, cluster, namespace, privilege)
		},
	}
}
