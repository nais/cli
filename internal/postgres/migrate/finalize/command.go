package finalize

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func Run(ctx context.Context, applicationName, targetInstanceName, team, environment string, dryRun bool) error {
	cfg := config.Config{
		AppName: applicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(environment))
	cfg.Team = team
	clientSet, err := k8s.SetupClientGo(environment)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientSet, cfg, dryRun, false)
	if err := migrator.Finalize(ctx); err != nil {
		return fmt.Errorf("error cleaning up instance: %w", err)
	}

	return nil
}
