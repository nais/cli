package postgrescmd

import (
	"fmt"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func passwordRotateCommand() *cli.Command {
	return &cli.Command{
		Name:        "rotate",
		Usage:       "Rotate the Postgres database password",
		Description: "The rotation is both done in GCP and in the Kubernetes secret",
		ArgsUsage:   "appname",
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
			database := context.String("database")

			return postgres.RotatePassword(context.Context, appName, namespace, cluster, database)
		},
	}
}
