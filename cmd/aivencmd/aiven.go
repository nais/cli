package aivencmd

import "github.com/nais/cli/pkg/metrics"

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	metrics := metrics.GetMetrics()
	metrics.RecordSubcommandUsage("aiven")
	metrics.PushMetrics(metrics.PushgatewayURL)

	return &cli.Command{
		Name:  "aiven",
		Usage: "Command used for management of AivenApplication",
		Subcommands: []*cli.Command{
			createCommand(),
			getCommand(),
			tidyCommand(),
		},
	}
}
