package setup

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

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return migrate.BeforeSubCommands(ctx, cmd)
}

func Action(ctx context.Context, cmd *cli.Command) error {
	cfg := migrate.MakeConfig(cmd)

	cluster := cmd.String("context")
	tier := cmd.String("tier")
	diskAutoresize := cmd.Bool("disk-autoresize")
	diskSize := cmd.Int("disk-size")
	instanceType := cmd.String("type")
	namespace := cmd.String("namespace")

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

	migrator := migrate.NewMigrator(client, clientset, cfg, cmd.Bool("dry-run"), cmd.Bool("no-wait"))
	if err := migrator.Setup(ctx); err != nil {
		return fmt.Errorf("error setting up migration: %w", err)
	}

	return nil
}

func TierFlagAction(ctx context.Context, cmd *cli.Command, v string) error {
	if !strings.HasPrefix(v, "db-") {
		return fmt.Errorf("tier must start with `db-`")
	}
	return nil
}

func TypeFlagAction(ctx context.Context, cmd *cli.Command, v string) error {
	if !strings.HasPrefix(v, "POSTGRES_") {
		return fmt.Errorf("instance type must start with `POSTGRES_`")
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
