package migratecmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/urfave/cli/v2"
)

const (
	contextFlagName = "context"
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
			cleanupCommand(),
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

func makeConfig(cCtx *cli.Context) migrate.Config {
	appName := cCtx.Args().Get(0)
	namespace := cCtx.Args().Get(1)
	targetInstanceName := cCtx.Args().Get(2)

	return migrate.Config{
		AppName:   appName,
		Namespace: namespace,
		Target: migrate.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}
}

func confirmContinue() error {
	fmt.Print("\nAre you sure you want to continue (y/N): ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
		return fmt.Errorf("cancelled by user")
	}

	return nil
}
