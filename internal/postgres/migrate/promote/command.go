package promote

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

type Flags struct {
	*migrate.Flags
	NoWait bool
}

func Run(ctx context.Context, args migrate.Arguments, flags *Flags) error {
	cfg := config.Config{
		AppName: args.ApplicationName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(args.TargetInstanceName),
		},
	}

	client := k8s.SetupControllerRuntimeClient(k8s.WithKubeContext(flags.Context))
	cfg.Namespace = client.CurrentNamespace

	clientSet, err := k8s.SetupClientGo(flags.Context)
	if err != nil {
		return err
	}

	migrator := migrate.NewMigrator(client, clientSet, cfg, flags.DryRun, flags.NoWait)
	if err := migrator.Promote(ctx); err != nil {
		return fmt.Errorf("error promoting instance: %w", err)
	}

	return nil
}
