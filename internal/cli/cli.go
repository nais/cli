package cli

import (
	"context"

	aivencommand "github.com/nais/cli/internal/aiven/command"
	debugcommand "github.com/nais/cli/internal/debug/command"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/metrics"
	naisdevicecommand "github.com/nais/cli/internal/naisdevice/command"
	postgrescommand "github.com/nais/cli/internal/postgres/command"
	validatecommand "github.com/nais/cli/internal/validate/command"
	"github.com/urfave/cli/v3"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func Run(ctx context.Context, args []string) error {
	app := &cli.Command{
		Name:                   "nais",
		Usage:                  "A Nais cli",
		Description:            "Nais platform utility cli, respects consoledonottrack.com",
		Version:                version + "-" + commit,
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Commands: []*cli.Command{
			{
				Name:        "login",
				Usage:       "Login using Google Auth.",
				Description: "This is a wrapper around gcloud auth login --update-adc.",
				Action:      gcp.LoginCommand,
			},
			{
				Name:  "kubeconfig",
				Usage: "Create a kubeconfig file for connecting to available clusters",
				Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "overwrite",
						Usage:   "Overwrite existing kubeconfig data if conflicts are found",
						Aliases: []string{"o"},
					},
					&cli.BoolFlag{
						Name:    "clear",
						Usage:   "Clear existing kubeconfig before writing new data",
						Aliases: []string{"c"},
					},
					&cli.StringSliceFlag{
						Name:    "exclude",
						Usage:   "Exclude clusters from kubeconfig. Can be specified multiple times or as a comma separated list",
						Aliases: []string{"e"},
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
					},
				},
				Before: kubeconfig.Before,
				Action: kubeconfig.Action,
			},
			validatecommand.Validate(),
			debugcommand.Debug(),
			aivencommand.Aiven(),
			naisdevicecommand.Device(),
			postgrescommand.Postgres(),
		},
	}

	setDefaults(app)
	metrics.CollectCommandHistogram(ctx, app.Commands)
	return app.Run(ctx, args)
}

func setDefaults(c *cli.Command) {
	c.HideHelpCommand = true

	for i := range c.Commands {
		setDefaults(c.Commands[i])
	}
}
