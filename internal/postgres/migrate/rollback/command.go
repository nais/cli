package rollback

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return migrate.BeforeSubCommands(ctx, cmd)
}

func Action(ctx context.Context, cmd *cli.Command) error {
	cfg := migrate.MakeConfig(cmd)
	cluster := cmd.String("context")

	pterm.Println(cmd.Description)

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
	cfg.Namespace = client.CurrentNamespace

	clientset, err := k8s.SetupClientGo(cluster)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientset, cfg, cmd.Bool("dry-run"), false)
	if err := migrator.Rollback(ctx); err != nil {
		return fmt.Errorf("error rolling back instance: %w", err)
	}

	return nil
}
