package migrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

const (
	tierFlagName           = "tier"
	diskAutoresizeFlagName = "disk-autoresize"
	diskSizeFlagName       = "disk-size"
	typeFlagName           = "type"
)

func setup() *cli.Command {
	return &cli.Command{
		Name:        "setup",
		Usage:       "Make necessary setup for a new migration",
		UsageText:   "nais postgres migrate setup APP_NAME TARGET_INSTANCE_NAME",
		Description: "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
		Flags: []cli.Flag{
			namespaceFlag(),
			kubeConfigFlag(),
			dryRunFlag(),
			noWaitFlag(),
			&cli.StringFlag{
				Name:        tierFlagName,
				Usage:       "The `TIER` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_TIER"),
				DefaultText: "Source instance value",
				Action: func(ctx context.Context, cmd *cli.Command, v string) error {
					if !strings.HasPrefix(v, "db-") {
						return fmt.Errorf("tier must start with `db-`")
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        diskAutoresizeFlagName,
				Usage:       "Enable disk autoresize for the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_DISK_AUTORESIZE"),
				DefaultText: "Source instance value",
			},
			&cli.IntFlag{
				Name:        diskSizeFlagName,
				Usage:       "The `DISK_SIZE` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_DISKSIZE"),
				DefaultText: "Source instance value",
			},
			&cli.StringFlag{
				Name:        typeFlagName,
				Usage:       "The `TYPE` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_TYPE"),
				DefaultText: "Source instance value",
				Action: func(ctx context.Context, cmd *cli.Command, v string) error {
					if !strings.HasPrefix(v, "POSTGRES_") {
						return fmt.Errorf("instance type must start with `POSTGRES_`")
					}
					return nil
				},
			},
		},
		Before: beforeFunc,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := makeConfig(cmd)

			cluster := cmd.String(contextFlagName)
			tier := cmd.String(tierFlagName)
			diskAutoresize := cmd.Bool(diskAutoresizeFlagName)
			diskSize := cmd.Int(diskSizeFlagName)
			instanceType := cmd.String(typeFlagName)
			namespace := cmd.String(namespaceFlagName)

			pterm.Println(cmd.Description)
			cfg.Target.Tier = isSet(tier)
			cfg.Target.DiskAutoresize = isSetBool(diskAutoresize)
			cfg.Target.DiskSize = isSetInt(diskSize)
			cfg.Target.Type = isSet(instanceType)

			client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(cluster))
			cfg.Namespace = client.CurrentNamespace
			if namespace != "" {
				cfg.Namespace = namespace
			}

			clientset, err := k8s.SetupClientGo(cluster)
			if err != nil {
				return err
			}

			migrator := migrate.NewMigrator(client, clientset, cfg, cmd.Bool(dryRunFlagName), cmd.Bool(noWaitFlagName))
			if err := migrator.Setup(ctx); err != nil {
				return fmt.Errorf("error setting up migration: %w", err)
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

func isSetBool(autoresize bool) option.Option[bool] {
	if autoresize {
		return option.Some(true)
	}
	return option.None[bool]()
}

func isSetInt(v int) option.Option[int] {
	if v == 0 {
		return option.None[int]()
	}
	return option.Some(v)
}
