package postgresCmd

import (
	"fmt"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func usersListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List users in a Postgres database",
		ArgsUsage: "appname",
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 1 {
				return fmt.Errorf("missing name of app")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().First()

			namespace := context.String("namespace")
			cluster := context.String("context")
			database := context.String("database")

			return postgres.ListUsers(context.Context, appName, cluster, namespace, database)
		},
	}
}
