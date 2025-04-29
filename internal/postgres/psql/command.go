package psql

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 1 {
		metrics.AddOne(ctx, "postgres_missing_app_name_error_total")
		return ctx, fmt.Errorf("missing name of app")
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	appName := cmd.Args().First()

	namespace := cmd.String("namespace")
	cluster := cmd.String("context")
	verbose := cmd.Bool("verbose")

	return postgres.RunPSQL(ctx, appName, cluster, namespace, verbose)
}
