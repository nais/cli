package migratecmd

import (
	"fmt"

	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	"github.com/urfave/cli/v2"
)

const (
	contextFlagName = "context"
	dryRunFlagName  = "dry-run"
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

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func beforeFunc(cCtx *cli.Context) error {
	argCount := cCtx.NArg()
	switch argCount {
	case 0:
		return fmt.Errorf("missing name of app")
	case 1:
		return fmt.Errorf("missing namespace")
	case 2:
		return fmt.Errorf("missing target instance name")
	case 3:
		return nil
	}

	return fmt.Errorf("too many arguments")
}

func makeConfig(cCtx *cli.Context) config.Config {
	appName := cCtx.Args().Get(0)
	namespace := cCtx.Args().Get(1)
	targetInstanceName := cCtx.Args().Get(2)

	return config.Config{
		AppName:   appName,
		Namespace: namespace,
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
