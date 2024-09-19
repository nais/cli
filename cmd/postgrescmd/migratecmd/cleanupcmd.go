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

func cleanupCommand() *cli.Command {
	return &cli.Command{
		Name:        "cleanup",
		Usage:       "Clean up after migration",
		UsageText:   "nais postgres migrate cleanup APP_NAME NAMESPACE TARGET_INSTANCE_NAME",
		Description: "Cleanup will remove the source instance and associated resources after a successful migration.",
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

This will delete the old database instance. Rollback after this point is not possible.

Only proceed if you are sure that the migration was successful and that your application is working as expected.
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

			err := migrator.Cleanup(context.Background())
			if err != nil {
				log.Fatalf("error cleaning up instance: %s", err)
			}
			return nil
		},
	}
}
