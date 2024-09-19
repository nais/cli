package migratecmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/option"
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
			appName := cCtx.Args().Get(0)
			namespace := cCtx.Args().Get(1)
			targetInstanceName := cCtx.Args().Get(2)

			cluster := cCtx.String(contextFlagName)

			fmt.Println(cCtx.Command.Description)

			fmt.Printf(`
Cluster (uses current context if unset): %s

Application: %s
Namespace: %s
Target Instance: %s

Your application will not be able to reach the database during promotion.
The database will be unavailable for a short period of time while the promotion is in progress.
`, cluster, appName, namespace, targetInstanceName)

			fmt.Print("\nAre you sure you want to continue (y/N): ")
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
				return fmt.Errorf("cancelled by user")
			}

			cfg := migrate.Config{
				AppName:   appName,
				Namespace: namespace,
				Target: migrate.InstanceConfig{
					InstanceName: option.Some(targetInstanceName),
				},
			}

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			migrator := migrate.NewMigrator(client, cfg)

			err := migrator.Promote(context.Background())
			if err != nil {
				log.Fatalf("error promoting instance: %s", err)
			}
			return nil
		},
	}
}
