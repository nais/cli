package postgrescmd

import (
	"fmt"

	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func grantCommand() *cli.Command {
	return &cli.Command{
		Name:        "grant",
		Usage:       "Grant yourself access to a Postgres database",
		Description: "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
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
				return fmt.Errorf("missing name of app")
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			appName := context.Args().First()

			namespace := context.String("namespace")
			cluster := context.String("context")

			return postgres.GrantAndCreateSQLUser(context.Context, appName, cluster, namespace)
		},
	}
}
