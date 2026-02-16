package finalize

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func Run(ctx context.Context, applicationName, targetInstanceName string, flags *flag.MigrateFinalize) error {
	cfg := config.Config{
		AppName: applicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(string(flags.Environment)))
	team, err := flags.RequiredTeam()
	if err != nil {
		return err
	}
	cfg.Team = team

	clientSet, err := k8s.SetupClientGo(string(flags.Environment))
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientSet, cfg, flags.DryRun, false)
	if err := migrator.Finalize(ctx); err != nil {
		return fmt.Errorf("error cleaning up instance: %w", err)
	}

	return nil
}
