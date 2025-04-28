package command

import (
	"context"
	"fmt"
	"slices"

	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func status() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Shows the status of your naisdevice",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Action: func(ctx context.Context, cmd *cli.Command, flag string) error {
					if !slices.Contains([]string{"yaml", "json"}, flag) {
						metrics.AddOne(ctx, "status_file_format_error_total")
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			outputFormat := cmd.String("output")
			quiet := cmd.Bool("quiet")
			verbose := cmd.Bool("verbose")

			status, err := naisdevice.GetStatus(ctx)
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
