package migratecmd

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

func promoteCommand() *cli.Command {
	return &cli.Command{
		Name:        "promote",
		Usage:       "Promote the migrated instance to the new primary instance",
		UsageText:   "nais postgres migrate promote APP_NAME TARGET_INSTANCE_NAME",
		Description: "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Args:        true,
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
			dryRunFlag(),
			noWaitFlag(),
		},
		Before: beforeFunc,
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			cluster := cCtx.String(contextFlagName)

			pterm.Println(cCtx.Command.Description)

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace

			migrator := migrate.NewMigrator(client, cfg, cCtx.Bool(dryRunFlagName), cCtx.Bool(noWaitFlagName))

			err := migrator.Promote(context.Background())
			if err != nil {
				return fmt.Errorf("error promoting instance: %w", err)
			}
			return nil
		},
	}
}
