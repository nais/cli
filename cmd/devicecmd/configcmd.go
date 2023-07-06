package devicecmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nais/cli/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"k8s.io/utils/strings/slices"
)

func configCommand() *cli.Command {
	return &cli.Command{
		Name:            "config",
		Usage:           "Adjust or view the naisdevice configuration",
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			getConfigCommand(),
			setConfigCommand(),
		},
	}
}

func getConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Gets the current configuration",
		Action: func(context *cli.Context) error {
			config, err := naisdevice.GetConfiguration(context.Context)
			if err != nil {
				return err
			}

			fmt.Printf("AutoConnect:\t%v\n", config.AutoConnect)
			fmt.Printf("CertRenewal:\t%v\n", config.CertRenewal)

			return nil
		},
	}
}

func setConfigCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Sets a configuration value",
		ArgsUsage: "setting value",
		Before: func(context *cli.Context) error {
			if context.Args().Len() >= 2 {
				return fmt.Errorf("missing required arguments: setting, value")
			}

			setting := strings.ToLower(context.Args().Get(0))
			value := context.Args().Get(1)

			if !slices.Contains(naisdevice.AllowedSettingsLowerCase, setting) {
				return fmt.Errorf("%v is not one of the allowed settings: %v", setting, strings.Join(naisdevice.AllowedSettings, ", "))
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

			return naisdevice.SetConfiguration(context.Context, setting, value)
		},
	}
}
