package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
	"k8s.io/utils/strings/slices"
)

func config() *cli.Command {
	return &cli.Command{
		Name:            "config",
		Usage:           "Adjust or view the naisdevice configuration",
		HideHelpCommand: true,
		Commands: []*cli.Command{
			getConfig(),
			setConfig(),
		},
	}
}

func getConfig() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Gets the current configuration",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := naisdevice.GetConfiguration(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("AutoConnect:\t%v\n", config.AutoConnect)

			return nil
		},
	}
}

func setConfig() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Sets a configuration value",
		ArgsUsage: "setting value",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Args().Len() < 2 {
				metrics.AddOne(ctx, "device_settings_error_total")
				return ctx, fmt.Errorf("missing required arguments: setting, value")
			}

			setting := cmd.Args().Get(0)
			value := cmd.Args().Get(1)
			if !slices.Contains(naisdevice.GetAllowedSettings(true, true), strings.ToLower(setting)) {
				metrics.AddOne(ctx, "device_settings_error_total")
				return ctx, fmt.Errorf("%v is not one of the allowed settings: %v", setting, strings.Join(naisdevice.GetAllowedSettings(false, false), ", "))
			}

			if _, err := strconv.ParseBool(value); err != nil {
				return ctx, err
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			setting := cmd.Args().Get(0)
			valueString := cmd.Args().Get(1)

			value, err := strconv.ParseBool(valueString)
			if err != nil {
				return err
			}

			if err := naisdevice.SetConfiguration(ctx, setting, value); err != nil {
				return err
			}

			fmt.Printf("%v has been set to %v\n", setting, value)

			return nil
		},
	}
}
