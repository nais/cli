package migratecmd

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

func rollbackCommand() *cli.Command {
	return &cli.Command{
		Name:        "rollback",
		Usage:       "Roll back the migration",
		UsageText:   "nais postgres migrate rollback APP_NAME TARGET_INSTANCE_NAME",
		Description: "Rollback will roll back the migration, and restore the application to use the original instance.",
		Args:        true,
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
			dryRunFlag(),
		},
		Before: beforeFunc,
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			cluster := cCtx.String(contextFlagName)

			pterm.Println(cCtx.Command.Description)

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace

			migrator := migrate.NewMigrator(client, cfg, cCtx.Bool(dryRunFlagName))

			err := migrator.Rollback(context.Background())
			if err != nil {
				return fmt.Errorf("error rolling back instance: %w", err)
			}
			return nil
		},
	}
}
