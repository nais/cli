package finalize

import (
	"context"
	"fmt"

	"github.com/nais/cli/v2/internal/k8s"
	"github.com/nais/cli/v2/internal/option"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/cli/v2/internal/postgres/migrate"
	"github.com/nais/cli/v2/internal/postgres/migrate/config"
)

func Run(ctx context.Context, applicationName, targetInstanceName string, flags *flag.MigrateFinalize) error {
	cfg := config.Config{
		AppName: applicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(string(flags.Context)))
	cfg.Namespace = flag.Namespace(client.CurrentNamespace)

	clientSet, err := k8s.SetupClientGo(string(flags.Context))
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientSet, cfg, flags.DryRun, false)
	if err := migrator.Finalize(ctx); err != nil {
		return fmt.Errorf("error cleaning up instance: %w", err)
	}

	return nil
}
