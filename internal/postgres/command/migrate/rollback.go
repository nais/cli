package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

func rollback() *cli.Command {
	return &cli.Command{
		Name:              "rollback",
		Usage:             "Roll back the migration",
		UsageText:         "nais postgres migrate rollback APP_NAME TARGET_INSTANCE_NAME",
		Description:       "Rollback will roll back the migration, and restore the application to use the original instance.",
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
			if err := migrator.Rollback(context.Background()); err != nil {
				return fmt.Errorf("error rolling back instance: %w", err)
			}

			return nil
		},
	}
}
