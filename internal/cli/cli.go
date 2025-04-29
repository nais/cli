package cli

import (
	"context"

	aivencreate "github.com/nais/cli/internal/aiven/create"
	aivenget "github.com/nais/cli/internal/aiven/get"
	aiventidy "github.com/nais/cli/internal/aiven/tidy"
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/tidy"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/metrics"
	naisdevicecommand "github.com/nais/cli/internal/naisdevice/command"
	postgrescommand "github.com/nais/cli/internal/postgres/command"
	"github.com/nais/cli/internal/validate"
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
			{
				Name:      "validate",
				Usage:     "Validate nais.yaml configuration",
				ArgsUsage: "nais.yaml [naiser.yaml...]",
				UsageText: "nais validate nais.yaml [naiser.yaml...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "vars",
						Usage: "path to `FILE` containing template variables, must be JSON or YAML format.",
					},
					&cli.StringSliceFlag{
						Name:  "var",
						Usage: "template variable in KEY=VALUE form, can be specified multiple times.",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "print all the template variables and final resources after templating.",
					},
				},
				Before: validate.Before,
				Action: validate.Action,
			},
			{
				Name:      "debug",
				Usage:     "Create and attach to a debug container",
				ArgsUsage: "workloadname",
				Description: "Create and attach to a debug pod or container. \n" +
					"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
					"allowing you to troubleshoot without affecting the live pod.\n" +
					"To debug a live pod, run the command without the '--copy' flag.\n" +
					"You can only reconnect to the debug session if the pod is running.",
				Commands: []*cli.Command{
					{
						Name:        "tidy",
						Usage:       "Clean up debug containers and debug pods",
						Description: "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
						ArgsUsage:   "workloadname",
						Flags: []cli.Flag{
							contextFlag(),
							namespaceFlag(),
							copyFlag(),
						},
						Before: tidy.Before,
						Action: tidy.Action,
					},
				},
				Flags: []cli.Flag{
					contextFlag(),
					copyFlag(),
					namespaceFlag(),
					byPodFlag(),
				},
				Before: debug.Before,
				Action: debug.Action,
			},
			{
				Name:  "aiven",
				Usage: "Command used for management of AivenApplication",
				Commands: []*cli.Command{
					{
						Name:      "create",
						Usage:     "Creates a protected and time-limited AivenApplication",
						ArgsUsage: "service username namespace",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:    "expire",
								Aliases: []string{"e"},
								Value:   1,
							},
							&cli.StringFlag{
								Name:    "pool",
								Aliases: []string{"p"},
								Value:   "nav-dev",
								Action:  aivencreate.PoolFlagAction,
							},
							&cli.StringFlag{
								Name:    "secret",
								Aliases: []string{"s"},
							},
							&cli.StringFlag{
								Name:    "instance",
								Aliases: []string{"i"},
								Action:  aivencreate.InstanceFlagAction,
							},
							&cli.StringFlag{
								Name:    "access",
								Aliases: []string{"a"},
								Action:  aivencreate.AccessFlagAction,
							},
						},
						Before: aivencreate.Before,
						Action: aivencreate.Action,
					},
					{
						Name:      "get",
						Usage:     "Generate preferred config format to '/tmp' folder",
						ArgsUsage: "service username namespace",
						Before:    aivenget.Before,
						Action:    aivenget.Action,
					},
					{
						Name:  "tidy",
						Usage: "Clean up /tmp/aiven-secret-* made by nais-cli",
						Description: `Remove '/tmp' folder '$TMPDIR' and files created by the aiven command
Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'`,
						Action: aiventidy.Action,
					},
				},
			},
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
