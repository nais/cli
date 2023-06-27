package postgresCmd

import (
	"fmt"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func usersAddCommand() *cli.Command {
	return &cli.Command{
		Name:    "add",
		Aliases: []string{"a"},
		Description: `Add user to a Postgres database.

Will grant user access to tables in public schema.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "privilege",
				Value: "select",
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 3 {
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
			database := context.String("database")
			privilege := context.String("privilege")

			return postgres.AddUser(context.Context, appName, username, password, cluster, namespace, database, privilege)
		},
	}
}
