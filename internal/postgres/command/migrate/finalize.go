package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

func finalize() *cli.Command {
	return &cli.Command{
		Name:              "finalize",
		Usage:             "Finalize the migration",
		UsageText:         "nais postgres migrate finalize APP_NAME TARGET_INSTANCE_NAME",
		Description:       "Finalize will remove the source instance and associated resources after a successful migration.",
		ReadArgsFromStdin: true, // TODO: Not sure about this one. Used to be `Args: true`, but field no longer exists
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
			dryRunFlag(),
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

			migrator := migrate.NewMigrator(client, clientset, cfg, cmd.Bool(dryRunFlagName), false)

			err = migrator.Finalize(context.Background())
			if err != nil {
				return fmt.Errorf("error cleaning up instance: %w", err)
			}
			return nil
		},
	}
}
