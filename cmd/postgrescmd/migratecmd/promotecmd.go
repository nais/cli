package migratecmd

import (
	"context"
	"fmt"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

func promoteCommand() *cli.Command {
	return &cli.Command{
		Name:        "promote",
		Usage:       "Promote the migrated instance to the new primary instance",
		UsageText:   "nais postgres migrate promote APP_NAME NAMESPACE TARGET_INSTANCE_NAME",
		Description: "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Args:        true,
		Flags: []cli.Flag{
			kubeConfigFlag(),
		},
		Before: beforeFunc,
		Action: func(cCtx *cli.Context) error {
			cfg := makeConfig(cCtx)
			cluster := cCtx.String(contextFlagName)

			fmt.Println(cCtx.Command.Description)
			fmt.Printf(`
Cluster (uses current context if unset): %s

Application: %s
Namespace: %s
Target Instance: %s

Your application will not be able to reach the database during promotion.
The database will be unavailable for a short period of time while the promotion is in progress.
`, cluster, cfg.AppName, cfg.Namespace, cfg.Target.InstanceName)

			err := confirmContinue()
			if err != nil {
				return err
			}

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			migrator := migrate.NewMigrator(client, cfg)

			err = migrator.Promote(context.Background())
			if err != nil {
				return fmt.Errorf("error promoting instance: %w", err)
			}
			return nil
		},
	}
}