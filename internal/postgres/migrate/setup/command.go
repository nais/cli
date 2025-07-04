package setup

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func Run(ctx context.Context, applicationName, targetInstanceName string, flags *flag.MigrateSetup) error {
	cfg := config.Config{
		AppName: applicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}

	cluster := flags.Context
	tier := flags.Tier
	diskAutoresize := flags.DiskAutoResize
	diskSize := flags.DiskSize
	instanceType := flags.InstanceType
	namespace := flags.Namespace

	cfg.Target.Tier = isSet(tier)
	cfg.Target.DiskAutoresize = isSetBool(diskAutoresize)
	cfg.Target.DiskSize = isSetInt(diskSize)
	cfg.Target.Type = isSet(instanceType)

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(string(cluster)))
	cfg.Namespace = flag.Namespace(client.CurrentNamespace)
	if namespace != "" {
		cfg.Namespace = namespace
	}

	clientSet, err := k8s.SetupClientGo(string(cluster))
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientSet, cfg, flags.DryRun, flags.NoWait)
	if err := migrator.Setup(ctx); err != nil {
		return fmt.Errorf("error setting up migration: %w", err)
	}

	return nil
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
