package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/metrics"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v2"
	"k8s.io/utils/strings/slices"
)

func config() *cli.Command {
	return &cli.Command{
		Name:            "config",
		Usage:           "Adjust or view the naisdevice configuration",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			getConfig(),
			setConfig(),
		},
	}
}

func getConfig() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Gets the current configuration",
		Action: func(context *cli.Context) error {
			config, err := naisdevice.GetConfiguration(context.Context)
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
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 2 {
				metrics.AddOne("device_settings_error_total")
				return fmt.Errorf("missing required arguments: setting, value")
			}

			setting := context.Args().Get(0)
			value := context.Args().Get(1)
			if !slices.Contains(naisdevice.GetAllowedSettings(true, true), strings.ToLower(setting)) {
				metrics.AddOne("device_settings_error_total")
				return fmt.Errorf("%v is not one of the allowed settings: %v", setting, strings.Join(naisdevice.GetAllowedSettings(false, false), ", "))
			}

			_, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			setting := context.Args().Get(0)
			valueString := context.Args().Get(1)

			value, err := strconv.ParseBool(valueString)
			if err != nil {
				return err
			}

			err = naisdevice.SetConfiguration(context.Context, setting, value)
			if err != nil {
				return err
			}

			fmt.Printf("%v has been set to %v\n", setting, value)

			return nil
		},
	}
}
