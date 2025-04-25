package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func revoke() *cli.Command {
	return &cli.Command{
		Name:  "revoke",
		Usage: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
		Description: `Revoke will revoke the role 'cloudsqliamuser' access to the
tables in the postgres instance. This is done by connecting using the application
credentials and modify the permissions on the public schema.

This operation is only required to run once for each postgresql instance.`,
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
			&cli.StringFlag{
				Name:  "schema",
				Value: "public",
				Usage: "Schema to revoke access from",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 1 {
				metrics.AddOne(ctx, "postgres_missing_app_name_error_total")
				return ctx, fmt.Errorf("missing name of app")
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			appName := cmd.Args().First()

			namespace := cmd.String("namespace")
			cluster := cmd.String("context")
			schema := cmd.String("schema")

			fmt.Println(cmd.Description)

			fmt.Print("\nAre you sure you want to continue (y/N): ")
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
				return fmt.Errorf("cancelled by user")
			}

			return postgres.RevokeAccess(ctx, appName, namespace, cluster, schema)
		},
	}
}
