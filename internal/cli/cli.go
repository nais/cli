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
	naisdeviceconfigget "github.com/nais/cli/internal/naisdevice/config/get"
	naisdeviceconfigset "github.com/nais/cli/internal/naisdevice/config/set"
	naisdeviceconnect "github.com/nais/cli/internal/naisdevice/connect"
	naisdevicedisconnect "github.com/nais/cli/internal/naisdevice/disconnect"
	naisdevicedoctor "github.com/nais/cli/internal/naisdevice/doctor"
	naisdevicejita "github.com/nais/cli/internal/naisdevice/jita"
	naisdevicestatus "github.com/nais/cli/internal/naisdevice/status"
	"github.com/nais/cli/internal/postgres"
	postgresaudit "github.com/nais/cli/internal/postgres/audit"
	"github.com/nais/cli/internal/postgres/command/migrate"
	postgresgrant "github.com/nais/cli/internal/postgres/grant"
	postgrespasswordrotate "github.com/nais/cli/internal/postgres/password/rotate"
	postgresprepare "github.com/nais/cli/internal/postgres/prepare"
	postgresproxy "github.com/nais/cli/internal/postgres/proxy"
	postgrespsql "github.com/nais/cli/internal/postgres/psql"
	postgresrevoke "github.com/nais/cli/internal/postgres/revoke"
	postgresusersadd "github.com/nais/cli/internal/postgres/users/add"
	postgresuserslist "github.com/nais/cli/internal/postgres/users/list"
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
			{
				Name:  "device",
				Usage: "Command used for management of naisdevice",
				Commands: []*cli.Command{
					{
						Name:  "config",
						Usage: "Adjust or view the naisdevice configuration",
						Commands: []*cli.Command{
							{
								Name:   "get",
								Usage:  "Gets the current configuration",
								Action: naisdeviceconfigget.Action,
							},
							{
								Name:      "set",
								Usage:     "Sets a configuration value",
								ArgsUsage: "setting value",
								Before:    naisdeviceconfigset.Before,
								Action:    naisdeviceconfigset.Action,
							},
						},
					},
					{
						Name:   "connect",
						Usage:  "Creates a naisdevice connection, will lock until connection",
						Action: naisdeviceconnect.Action,
					},
					{
						Name:   "disconnect",
						Usage:  "Disconnects your naisdevice",
						Action: naisdevicedisconnect.Action,
					},
					{
						Name:      "jita",
						Usage:     "Connects to a JITA gateway",
						ArgsUsage: "gateway",
						Before:    naisdevicejita.Before,
						Action:    naisdevicejita.Action,
					},
					{
						Name:  "status",
						Usage: "Shows the status of your naisdevice",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
								Action:  naisdevicestatus.OutputFlagAction,
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
						Action: naisdevicestatus.Action,
					},
					{
						Name:   "doctor",
						Usage:  "Examine the health of your naisdevice",
						Action: naisdevicedoctor.Action,
					},
				},
			},
			{
				Name:   "postgres",
				Usage:  "Command used for connecting to Postgres",
				Before: postgres.Before,
				Commands: []*cli.Command{
					{
						Name:        "enable-audit",
						Usage:       "Enable audit extension in Postgres database",
						Description: "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
						ArgsUsage:   "appname",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
						},
						Before: postgresaudit.Before,
						Action: postgresaudit.Action,
					},
					{
						Name:        "grant",
						Usage:       "Grant yourself access to a Postgres database",
						Description: "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
						ArgsUsage:   "appname",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
						},
						Before: postgresgrant.Before,
						Action: postgresgrant.Action,
					},
					migrate.Migrate(),
					{
						Name:  "password",
						Usage: "Administrate Postgres password",
						Commands: []*cli.Command{
							{
								Name:        "rotate",
								Usage:       "Rotate the Postgres database password",
								Description: "The rotation is both done in GCP and in the Kubernetes secret",
								ArgsUsage:   "appname",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "context",
										Aliases: []string{"c"},
									},
									&cli.StringFlag{
										Name:    "namespace",
										Aliases: []string{"n"},
									},
								},
								Before: postgrespasswordrotate.Before,
								Action: postgrespasswordrotate.Action,
							},
						},
					},
					{
						Name:  "prepare",
						Usage: "Prepare your postgres instance for use with personal accounts",
						Description: `Prepare will prepare the postgres instance by connecting using the
application credentials and modify the permissions on the public schema.
All IAM users in your GCP project will be able to connect to the instance.

This operation is only required to run once for each postgresql instance.`,
						ArgsUsage: "appname",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all-privs",
								Usage: "Gives all privileges to users",
							},
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
							&cli.StringFlag{
								Name:  "schema",
								Value: "public",
								Usage: "Schema to grant access to",
							},
						},
						Before: postgresprepare.Before,
						Action: postgresprepare.Action,
					},
					{
						Name:        "proxy",
						Usage:       "Create a proxy to a Postgres instance",
						Description: "Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.",
						ArgsUsage:   "appname",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:    "port",
								Aliases: []string{"p"},
								Value:   5432,
							},
							&cli.StringFlag{
								Name:    "host",
								Aliases: []string{"H"},
								Value:   "localhost",
							},
							&cli.BoolFlag{
								Name:    "verbose",
								Aliases: []string{"v"},
							},
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
						},
						Before: postgresproxy.Before,
						Action: postgresproxy.Action,
					},
					{
						Name:        "psql",
						Usage:       "Connect to the database using psql",
						Description: "Create a shell to the postgres instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
						ArgsUsage:   "appname",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "verbose",
								Aliases: []string{"v"},
							},
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
						},
						Before: postgrespsql.Before,
						Action: postgrespsql.Action,
					},
					{
						Name:  "revoke",
						Usage: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
						Description: `Revoke will revoke the role 'cloudsqliamuser' access to the
tables in the postgres instance. This is done by connecting using the application
credentials and modify the permissions on the public schema.

This operation is only required to run once for each postgresql instance.`,
						ArgsUsage: "appname",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "context",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:    "namespace",
								Aliases: []string{"n"},
							},
							&cli.StringFlag{
								Name:  "schema",
								Value: "public",
								Usage: "Schema to revoke access from",
							},
						},
						Before: postgresrevoke.Before,
						Action: postgresrevoke.Action,
					},
					{
						Name:        "users",
						Usage:       "Administrate users in your Postgres instance",
						Description: "Command used for listing and adding users to database",
						Commands: []*cli.Command{
							{
								Name:        "add",
								Usage:       "Add user to a Postgres database",
								Description: "Will grant a user access to tables in public schema.",
								ArgsUsage:   "appname username password",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "privilege",
										Aliases: []string{"p"},
										Value:   "select",
									},
									&cli.StringFlag{
										Name:    "context",
										Aliases: []string{"c"},
									},
									&cli.StringFlag{
										Name:    "namespace",
										Aliases: []string{"n"},
									},
								},
								Before: postgresusersadd.Before,
								Action: postgresusersadd.Action,
							},
							{
								Name:      "list",
								Usage:     "List users in a Postgres database",
								ArgsUsage: "appname",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:    "context",
										Aliases: []string{"c"},
									},
									&cli.StringFlag{
										Name:    "namespace",
										Aliases: []string{"n"},
									},
								},
								Before: postgresuserslist.Before,
								Action: postgresuserslist.Action,
							},
						},
					},
				},
			},
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
