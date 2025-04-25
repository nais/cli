package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func usersAdd() *cli.Command {
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 3 {
				metrics.AddOne(ctx, "postgres_missing_args_error_total")
				return ctx, fmt.Errorf("missing required arguments: appname, username, password")
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			appName := cmd.Args().Get(0)
			username := cmd.Args().Get(1)
			password := cmd.Args().Get(2)

			namespace := cmd.String("namespace")
			cluster := cmd.String("context")
			privilege := cmd.String("privilege")

			return postgres.AddUser(ctx, appName, username, password, cluster, namespace, privilege)
		},
	}
}
