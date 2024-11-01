package devicecmd

import (
	"fmt"

	"github.com/nais/cli/pkg/metrics"
	"github.com/nais/cli/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"k8s.io/utils/strings/slices"
)

func statusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Shows the status of your naisdevice",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Action: func(context *cli.Context, flag string) error {
					if !slices.Contains([]string{"yaml", "json"}, flag) {
						metrics.AddOne("status_file_format_error_total")
						return fmt.Errorf("%v is not an implemented format", flag)
					}

					return nil
				},
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Action: func(context *cli.Context) error {
			metrics.AddOne("device_status_total")
			outputFormat := context.String("output")
			quiet := context.Bool("quiet")
			verbose := context.Bool("verbose")

			status, err := naisdevice.GetStatus(context.Context)
			if err != nil {
				return err
			}

			if quiet {
				if !naisdevice.IsConnected(status) {
					return cli.Exit("", 1)
				}
				return nil
			}

			if outputFormat != "" {
				return naisdevice.PrintFormattedStatus(outputFormat, status)
			}

			if verbose {
				naisdevice.PrintVerboseStatus(status)
				return nil
			}

			fmt.Println(status.ConnectionState.String())

			return nil
		},
	}
}
