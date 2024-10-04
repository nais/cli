package migratecmd

import (
	"context"
	"fmt"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

func finalizeCommand() *cli.Command {
	return &cli.Command{
		Name:        "finalize",
		Usage:       "Finalize the migration",
		UsageText:   "nais postgres migrate finalize APP_NAME NAMESPACE TARGET_INSTANCE_NAME",
		Description: "Finalize will remove the source instance and associated resources after a successful migration.",
		Args:        true,
		Flags: []cli.Flag{
			kubeConfigFlag(),
			dryRunFlag(),
		},
		Before: beforeFunc,
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			cluster := cCtx.String(contextFlagName)

			fmt.Println(cCtx.Command.Description)

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			migrator := migrate.NewMigrator(client, cfg, cCtx.Bool(dryRunFlagName))

			err := migrator.Finalize(context.Background())
			if err != nil {
				return fmt.Errorf("error cleaning up instance: %w", err)
			}
			return nil
		},
	}
}
