package prepare

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

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 1 {
		metrics.AddOne(ctx, "postgres_prepare_missing_app_name_error_total")
		return ctx, fmt.Errorf("missing name of app")
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	appName := cmd.Args().First()

	allPrivs := cmd.Bool("all-privs")
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

	return postgres.PrepareAccess(ctx, appName, namespace, cluster, schema, allPrivs)
}
