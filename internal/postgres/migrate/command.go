package migrate

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate/config"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	_, err := gcp.ValidateAndGetUserLogin(ctx, false)
	return ctx, err
}

func BeforeSubCommands(ctx context.Context, cmd *cli.Command) (context.Context, error) {
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

func MakeConfig(cmd *cli.Command) config.Config {
	appName := cmd.Args().Get(0)
	targetInstanceName := cmd.Args().Get(1)

	return config.Config{
		AppName: appName,
		Target: config.InstanceConfig{
			InstanceName: option.Some(targetInstanceName),
		},
	}
}
