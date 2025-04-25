package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate/config"
	"github.com/urfave/cli/v3"
)

const (
	namespaceFlagName = "namespace"
	contextFlagName   = "context"
	dryRunFlagName    = "dry-run"
	noWaitFlagName    = "no-wait"
)

func Migrate() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Command used for migrating to a new Postgres instance",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			_, err := gcp.ValidateAndGetUserLogin(ctx, false)
			return ctx, err
		},
		Commands: []*cli.Command{
			setup(),
			promote(),
			finalize(),
			rollback(),
		},
	}
}

func namespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        namespaceFlagName,
		DefaultText: "The namespace from your current kubeconfig context",
		Usage:       "The kubernetes `NAMESPACE` to use",
		Aliases:     []string{"n"},
	}
}

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func noWaitFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  noWaitFlagName,
		Usage: "Do not wait for the job to complete",
	}
}

func beforeFunc(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	argCount := cmd.NArg()
	switch argCount {
	case 0:
		return ctx, fmt.Errorf("missing name of app")
	case 1:
		return ctx, fmt.Errorf("missing target instance name")
	case 2:
		return ctx, nil
	}

	return ctx, fmt.Errorf("too many arguments")
}

func makeConfig(cmd *cli.Command) config.Config {
	appName := cmd.Args().Get(0)
	targetInstanceName := cmd.Args().Get(1)

	return config.Config{
		AppName: appName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}
}

func dryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  dryRunFlagName,
		Usage: "Perform a dry run of the migration setup, without actually starting the migration",
	}
}
