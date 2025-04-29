package add

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/postgres"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 3 {
		metrics.AddOne(ctx, "postgres_missing_args_error_total")
		return ctx, fmt.Errorf("missing required arguments: appname, username, password")
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	appName := cmd.Args().Get(0)
	username := cmd.Args().Get(1)
	password := cmd.Args().Get(2)

	namespace := cmd.String("namespace")
	cluster := cmd.String("context")
	privilege := cmd.String("privilege")

	return postgres.AddUser(ctx, appName, username, password, cluster, namespace, privilege)
}
