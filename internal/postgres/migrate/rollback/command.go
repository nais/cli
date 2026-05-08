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

func Run(ctx context.Context, applicationName, targetInstanceName, team, environment string, flags *flag.MigrateRollback) error {
	cfg := config.Config{
		AppName: applicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(environment))
	cfg.Team = team
	clientset, err := k8s.SetupClientGo(environment)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientset, cfg, flags.DryRun, false)
	if err := migrator.Rollback(ctx); err != nil {
		return fmt.Errorf("error rolling back instance: %w", err)
	}

	return nil
}
