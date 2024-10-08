package migratecmd

import (
	"context"
	"fmt"
	"github.com/pterm/pterm"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

func finalizeCommand() *cli.Command {
	return &cli.Command{
		Name:        "finalize",
		Usage:       "Finalize the migration",
		UsageText:   "nais postgres migrate finalize APP_NAME TARGET_INSTANCE_NAME",
		Description: "Finalize will remove the source instance and associated resources after a successful migration.",
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

			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace

			clientset, err := k8s.SetupClientGo(cluster)
			if err != nil {
				return err
			}

			migrator := migrate.NewMigrator(client, clientset, cfg, cCtx.Bool(dryRunFlagName), false)

			err = migrator.Finalize(context.Background())
			if err != nil {
				return fmt.Errorf("error cleaning up instance: %w", err)
			}
			return nil
		},
	}
}
