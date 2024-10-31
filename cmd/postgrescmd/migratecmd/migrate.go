package migratecmd

import (
	"fmt"

	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	"github.com/urfave/cli/v2"
)

const (
	namespaceFlagName = "namespace"
	contextFlagName   = "context"
	dryRunFlagName    = "dry-run"
	noWaitFlagName    = "no-wait"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Command used for migrating to a new Postgres instance",
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context, false)
		},
		Subcommands: []*cli.Command{
			setupCommand(),
			promoteCommand(),
			finalizeCommand(),
			rollbackCommand(),
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

func beforeFunc(cCtx *cli.Context) error {
	argCount := cCtx.NArg()
	switch argCount {
	case 0:
		return fmt.Errorf("missing name of app")
	case 1:
		return fmt.Errorf("missing target instance name")
	case 2:
		return nil
	}

	return fmt.Errorf("too many arguments")
}

func makeConfig(cCtx *cli.Context) config.Config {
	appName := cCtx.Args().Get(0)
	targetInstanceName := cCtx.Args().Get(1)

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
