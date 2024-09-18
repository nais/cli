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

const (
	tierFlagName     = "tier"
	diskSizeFlagName = "disk-size"
	typeFlagName     = "type"
)

func setupCommand() *cli.Command {
	return &cli.Command{
		Name:        "setup",
		Usage:       "Make necessary setup for a new migration",
		UsageText:   "nais postgres migrate setup APP_NAME NAMESPACE TARGET_INSTANCE_NAME",
		Description: "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
		Args:        true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        contextFlagName,
				Aliases:     []string{"c"},
				Usage:       "The kubeconfig `CONTEXT` to use",
				DefaultText: "The current context in your kubeconfig",
			},
			&cli.StringFlag{
				Name:        tierFlagName,
				Usage:       "The `TIER` of the new instance",
				Category:    "Target instance configuration",
				EnvVars:     []string{"TARGET_INSTANCE_TIER"},
				DefaultText: "Source instance value",
				Action: func(context *cli.Context, v string) error {
					if !strings.HasPrefix(v, "db-") {
						return fmt.Errorf("tier must start with db-")
					}
					return nil
				},
			},
			&cli.IntFlag{
				Name:        diskSizeFlagName,
				Usage:       "The `DISK_SIZE` of the new instance",
				Category:    "Target instance configuration",
				EnvVars:     []string{"TARGET_INSTANCE_DISKSIZE"},
				DefaultText: "Source instance value",
			},
			&cli.StringFlag{
				Name:        typeFlagName,
				Usage:       "The `TYPE` of the new instance",
				Category:    "Target instance configuration",
				EnvVars:     []string{"TARGET_INSTANCE_TYPE"},
				DefaultText: "Source instance value",
				Action: func(context *cli.Context, v string) error {
					if !strings.HasPrefix(v, "POSTGRES_") {
						return fmt.Errorf("instance type must start with POSTGRES_")
					}
					return nil
				},
			},
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
			tier := cCtx.String(tierFlagName)
			diskSize := cCtx.Int(diskSizeFlagName)
			instanceType := cCtx.String(typeFlagName)

			fmt.Println(cCtx.Command.Description)

			fmt.Printf(`
Cluster (uses current context if unset): %s

Application: %s
Namespace: %s
Target Instance: %s

Optional configuration, if blank, keeps current value:
Tier: %s
Disk Size: %d
Instance Type: %s
`, cluster, appName, namespace, targetInstanceName, tier, diskSize, instanceType)

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
					Tier:         isSet(tier),
					DiskSize:     isSetInt(diskSize),
					Type:         isSet(instanceType),
				},
			}

			client := k8s.SetupClient(k8s.WithKubeContext(cluster))
			migrator := migrate.NewMigrator(client, cfg)

			err := migrator.Setup(context.Background())
			if err != nil {
				log.Fatalf("error setting up migration: %s", err)
			}
			return nil
		},
	}
}

func isSet(v string) option.Option[string] {
	if v == "" {
		return option.None[string]()
	}
	return option.Some(v)
}

func isSetInt(v int) option.Option[int] {
	if v == 0 {
		return option.None[int]()
	}
	return option.Some(v)
}
