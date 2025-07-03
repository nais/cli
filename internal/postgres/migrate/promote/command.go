package promote

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func Run(ctx context.Context, applicationName, targetInstanceName string, flags *flag.MigratePromote) error {
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

	migrator := migrate.NewMigrator(client, clientSet, cfg, flags.DryRun, flags.NoWait)
	if err := migrator.Promote(ctx); err != nil {
		return fmt.Errorf("error promoting instance: %w", err)
	}

	return nil
}
