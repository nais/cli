package rollback

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func Run(ctx context.Context, args migrate.Arguments, flags *flag.MigrateRollback) error {
	cfg := config.Config{
		AppName: args.ApplicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(args.TargetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(flags.Context))
	cfg.Namespace = client.CurrentNamespace

	clientset, err := k8s.SetupClientGo(flags.Context)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientset, cfg, flags.DryRun, false)
	if err := migrator.Rollback(ctx); err != nil {
		return fmt.Errorf("error rolling back instance: %w", err)
	}

	return nil
}
