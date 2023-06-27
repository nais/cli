package postgresCmd

import (
	"fmt"
	"github.com/nais/cli/pkg/postgres"
	"github.com/urfave/cli/v2"
)

func proxyCommand() *cli.Command {
	return &cli.Command{
		Name:        "proxy",
		Description: "Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
			&cli.UintFlag{
				Name:  "port",
				Value: 5432,
			},
			&cli.StringFlag{
				Name:  "host",
				Value: "localhost",
			},
		},
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
			verbose := context.Bool("verbose")
			port := context.Uint("port")
			host := context.String("host")

			return postgres.RunProxy(context.Context, appName, cluster, namespace, database, host, port, verbose)
		},
	}
}
