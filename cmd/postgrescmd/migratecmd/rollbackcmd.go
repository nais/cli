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

func rollbackCommand() *cli.Command {
	return &cli.Command{
		Name:        "rollback",
		Usage:       "Roll back the migration",
		UsageText:   "nais postgres migrate rollback APP_NAME NAMESPACE TARGET_INSTANCE_NAME",
		Description: "Rollback will roll back the migration, and restore the application to use the original instance.",
		Args:        true,
		Flags: []cli.Flag{
			kubeConfigFlag(),
		},
		Before: func(cCtx *cli.Context) error {
			argCount := cCtx.NArg()
			switch argCount {
			case 0:
				return fmt.Errorf("missing name of app")
			case 1:
				return fmt.Errorf("missing namespace")
			case 2:
				return fmt.Errorf("missing target instance name")
			case 3:
				return nil
			}

			return fmt.Errorf("too many arguments")
		},
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

This will roll back the migration, and restore the application to use the original instance.
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

			err := migrator.Rollback(context.Background())
			if err != nil {
				log.Fatalf("error rolling back instance: %s", err)
			}
			return nil
		},
	}
}
