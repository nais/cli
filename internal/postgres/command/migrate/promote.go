package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

func promote() *cli.Command {
	return &cli.Command{
		Name:        "promote",
		Usage:       "Promote the migrated instance to the new primary instance",
		UsageText:   "nais postgres migrate promote APP_NAME TARGET_INSTANCE_NAME",
		Description: "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
			dryRunFlag(),
			noWaitFlag(),
		},
		Before: beforeFunc,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := makeConfig(cmd)
			cluster := cmd.String(contextFlagName)

			pterm.Println(cmd.Description)

			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace

			clientset, err := k8s.SetupClientGo(cluster)
			if err != nil {
				return err
			}

			migrator := migrate.NewMigrator(client, clientset, cfg, cmd.Bool(dryRunFlagName), cmd.Bool(noWaitFlagName))
			if err := migrator.Promote(ctx); err != nil {
				return fmt.Errorf("error promoting instance: %w", err)
			}

			return nil
		},
	}
}
