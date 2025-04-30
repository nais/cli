package cli2

import (
	aivencreate "github.com/nais/cli/internal/aiven/create"
	aivenget "github.com/nais/cli/internal/aiven/get"
	aiventidy "github.com/nais/cli/internal/aiven/tidy"
	naisdeviceconfigget "github.com/nais/cli/internal/naisdevice/config/get"
	naisdeviceconfigset "github.com/nais/cli/internal/naisdevice/config/set"
	naisdeviceconnect "github.com/nais/cli/internal/naisdevice/connect"
	naisdevicedisconnect "github.com/nais/cli/internal/naisdevice/disconnect"
	naisdevicedoctor "github.com/nais/cli/internal/naisdevice/doctor"
	naisdevicejita "github.com/nais/cli/internal/naisdevice/jita"
	naisdevicestatus "github.com/nais/cli/internal/naisdevice/status"
	"github.com/nais/cli/internal/postgres"
	postgresaudit "github.com/nais/cli/internal/postgres/audit"
	postgresgrant "github.com/nais/cli/internal/postgres/grant"
	postgresmigrate "github.com/nais/cli/internal/postgres/migrate"
	postgresmigratefinalize "github.com/nais/cli/internal/postgres/migrate/finalize"
	postgresmigratepromote "github.com/nais/cli/internal/postgres/migrate/promote"
	postgresmigraterollback "github.com/nais/cli/internal/postgres/migrate/rollback"
	postgresmigratesetup "github.com/nais/cli/internal/postgres/migrate/setup"
	postgrespasswordrotate "github.com/nais/cli/internal/postgres/password/rotate"
	postgresprepare "github.com/nais/cli/internal/postgres/prepare"
	postgresproxy "github.com/nais/cli/internal/postgres/proxy"
	postgrespsql "github.com/nais/cli/internal/postgres/psql"
	postgresrevoke "github.com/nais/cli/internal/postgres/revoke"
	postgresusersadd "github.com/nais/cli/internal/postgres/users/add"
	postgresuserslist "github.com/nais/cli/internal/postgres/users/list"
	"github.com/spf13/cobra"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func Run() error {
	app := &cobra.Command{
		Use:        "nais",
		Short:      "A Nais cli",
		Long:       "Nais platform utility cli, respects consoledonottrack.com",
		Deprecated: "deprecated, use: this instead",
		Version:    version + "-" + commit,
	}

	loginCommand := &cobra.Command{
		Use:   "login",
		Short: "Login using Google Auth.",
		Long:  "This is a wrapper around gcloud auth login --update-adc.",
		// Run:   gcp.LoginCommand,
	}
	app.AddCommand(loginCommand)

	kubeconfigCommand := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Create a kubeconfig file for connecting to available clusters",
		Long: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		// Before: kubeconfig.Before,
		// Run:    kubeconfig.Action,
	}
	kubeconfigCommand.Flags().StringSlice("exlude", nil, "Exclude clusters from kubeconfig. Can be specified as a comma separated list")
	kubeconfigCommand.Flags().Bool("overwrite", false, "Overwrite existing kubeconfig data if conflicts are found")
	kubeconfigCommand.Flags().Bool("clear", false, "Clear existing kubeconfig before writing new data")
	kubeconfigCommand.Flags().Bool("verbose", false, "Verbose output")

	validateCommand := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate nais.yaml configuration",
		// Before: validate.Before,
		// Run:    validate.Action,
	}
	validateCommand.Flags().String("vars", "", "Path to the `file` containing template variables, must be JSON or YAML format.")
	validateCommand.Flags().StringArray("var", nil, "Template variable in KEY=VALUE form, can be specified multiple times.")
	validateCommand.Flags().String("verbose", "", "Print all the template variables and final resources after templating.")

	debugCommand := &cobra.Command{
		Use:   "debug [app]",
		Short: "Create and attach to a debug container for a given `app`",
		Long: "Create and attach to a debug pod or container. \n" +
			"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
			"allowing you to troubleshoot without affecting the live pod.\n" +
			"To debug a live pod, run the command without the '--copy' flag.\n" +
			"You can only reconnect to the debug session if the pod is running.",
		// Before: debug.Before,
		// Run:    debug.Action,
	}
	debugCommand.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	debugCommand.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	debugCommand.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")
	debugCommand.Flags().Bool("by-pod", false, "Attach to a specific `BY-POD` in a workload")

	debugTidyCommand := &cobra.Command{
		Use:   "tidy [app]",
		Short: "Clean up debug containers and debug pods",
		Long:  "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		// Before: tidy.Before,
		// Run:    tidy.Action,
	}
	debugTidyCommand.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	debugTidyCommand.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	debugTidyCommand.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")

	aivenCommand := &cobra.Command{
		Use:   "aiven",
		Short: "Command used for management of AivenApplication",
	}

	aivenCreateCommand := &cobra.Command{
		Use:       "create",
		Short:     "Creates a protected and time-limited AivenApplication",
		ArgsShort: "service username namespace",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Use:     "expire",
				Aliases: []string{"e"},
				Value:   1,
			},
			&cli.StringFlag{
				Use:     "pool",
				Aliases: []string{"p"},
				Value:   "nav-dev",
				Run:     aivencreate.PoolFlagAction,
			},
			&cli.StringFlag{
				Use:     "secret",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Use:     "instance",
				Aliases: []string{"i"},
				Run:     aivencreate.InstanceFlagAction,
			},
			&cli.StringFlag{
				Use:     "access",
				Aliases: []string{"a"},
				Run:     aivencreate.AccessFlagAction,
			},
		},
		Before: aivencreate.Before,
		Run:    aivencreate.Action,
	}
	aivenGetCommand := &cobra.Command{
		Use:       "get",
		Short:     "Generate preferred config format to '/tmp' folder",
		ArgsShort: "service username namespace",
		Before:    aivenget.Before,
		Run:       aivenget.Action,
	}
	aivenTidyCommand := &cobra.Command{
		Use:   "tidy",
		Short: "Clean up /tmp/aiven-secret-* made by nais-cli",
		Long: `Remove '/tmp' folder '$TMPDIR' and files created by the aiven command
Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'`,
		Run: aiventidy.Action,
	}
	deviceCommand := &cobra.Command{
		Use:   "device",
		Short: "Command used for management of naisdevice",
	}
	deviceConfigCommand := &cobra.Command{
		Use:   "config",
		Short: "Adjust or view the naisdevice configuration",
	}
	deviceConfigGetCommand := &cobra.Command{
		Use:   "get",
		Short: "Gets the current configuration",
		Run:   naisdeviceconfigget.Action,
	}
	deviceConfigSetCommand := &cobra.Command{
		Use:       "set",
		Short:     "Sets a configuration value",
		ArgsShort: "setting value",
		Before:    naisdeviceconfigset.Before,
		Run:       naisdeviceconfigset.Action,
	}
	deviceConnectCommand := &cobra.Command{
		Use:   "connect",
		Short: "Creates a naisdevice connection, will lock until connection",
		Run:   naisdeviceconnect.Action,
	}
	deviceDisconnectCommand := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnects your naisdevice",
		Run:   naisdevicedisconnect.Action,
	}
	deviceJitaCommand := &cobra.Command{
		Use:       "jita",
		Short:     "Connects to a JITA gateway",
		ArgsShort: "gateway",
		Before:    naisdevicejita.Before,
		Run:       naisdevicejita.Action,
	}
	deviceStatusCommand := &cobra.Command{
		Use:   "status",
		Short: "Shows the status of your naisdevice",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Use:     "output",
				Aliases: []string{"o"},
				Run:     naisdevicestatus.OutputFlagAction,
			},
			&cli.BoolFlag{
				Use:     "quiet",
				Aliases: []string{"q"},
			},
			&cli.BoolFlag{
				Use:     "verbose",
				Aliases: []string{"v"},
			},
		},
		Run: naisdevicestatus.Action,
	}
	deviceDoctorCommand := &cli.Command{
		Use:   "doctor",
		Short: "Examine the health of your naisdevice",
		Run:   naisdevicedoctor.Action,
	}
	postgresCommand := &cobra.Command{
		Use:    "postgres",
		Short:  "Command used for connecting to Postgres",
		Before: postgres.Before,
	}
	postgresEnableAuditCommand := &cli.Command{
		Use:       "enable-audit",
		Short:     "Enable audit extension in Postgres database",
		Long:      "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgresaudit.Before,
		Run:    postgresaudit.Action,
	}
	postgresGrandtCommand := &cli.Command{
		Use:       "grant",
		Short:     "Grant yourself access to a Postgres database",
		Long:      "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgresgrant.Before,
		Run:    postgresgrant.Action,
	}
	postgresMigrateCommand := &cli.Command{
		Use:    "migrate",
		Short:  "Command used for migrating to a new Postgres instance",
		Before: postgresmigrate.Before,
	}
	postgresMigrateSetupCommand := &cli.Command{
		Use:       "setup",
		Short:     "Make necessary setup for a new migration",
		UsageText: "nais postgres migrate setup APP_NAME TARGET_INSTANCE_NAME",
		Long:      "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
		Flags: []cli.Flag{
			namespaceFlag(),
			contextFlag(),
			dryRunFlag(),
			noWaitFlag(),
			&cli.StringFlag{
				Use:         "tier",
				Short:       "The `TIER` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_TIER"),
				DefaultText: "Source instance value",
				Run:         postgresmigratesetup.TierFlagAction,
			},
			&cli.BoolFlag{
				Use:         "disk-autoresize",
				Short:       "Enable disk autoresize for the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_DISK_AUTORESIZE"),
				DefaultText: "Source instance value",
			},
			&cli.IntFlag{
				Use:         "disk-size",
				Short:       "The `DISK_SIZE` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_DISKSIZE"),
				DefaultText: "Source instance value",
			},
			&cli.StringFlag{
				Use:         "type",
				Short:       "The `TYPE` of the new instance",
				Category:    "Target instance configuration",
				Sources:     cli.EnvVars("TARGET_INSTANCE_TYPE"),
				DefaultText: "Source instance value",
				Run:         postgresmigratesetup.TypeFlagAction,
			},
		},
		Before: postgresmigratesetup.Before,
		Run:    postgresmigratesetup.Action,
	}
	postgresMigratePromoteCommand := &cli.Command{
		Use:       "promote",
		Short:     "Promote the migrated instance to the new primary instance",
		UsageText: "nais postgres migrate promote APP_NAME TARGET_INSTANCE_NAME",
		Long:      "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Flags: []cli.Flag{
			namespaceFlag(),
			contextFlag(),
			dryRunFlag(),
			noWaitFlag(),
		},
		Before: postgresmigratepromote.Before,
		Run:    postgresmigratepromote.Action,
	}
	postgresMigrateFinalizeCommand := &cli.Command{
		Use:       "finalize",
		Short:     "Finalize the migration",
		UsageText: "nais postgres migrate finalize APP_NAME TARGET_INSTANCE_NAME",
		Long:      "Finalize will remove the source instance and associated resources after a successful migration.",
		Flags: []cli.Flag{
			namespaceFlag(),
			contextFlag(),
			dryRunFlag(),
		},
		Before: postgresmigratefinalize.Before,
		Run:    postgresmigratefinalize.Action,
	}
	postgresMigrateRollbackCommand := &cli.Command{
		Use:       "rollback",
		Short:     "Roll back the migration",
		UsageText: "nais postgres migrate rollback APP_NAME TARGET_INSTANCE_NAME",
		Long:      "Rollback will roll back the migration, and restore the application to use the original instance.",
		Flags: []cli.Flag{
			namespaceFlag(),
			contextFlag(),
			dryRunFlag(),
		},
		Before: postgresmigraterollback.Before,
		Run:    postgresmigraterollback.Action,
	}

	postgresPasswordCommand := &cli.Command{
		Use:   "password",
		Short: "Administrate Postgres password",
	}
	postgresPasswordRotateCommand := &cli.Command{
		Use:       "rotate",
		Short:     "Rotate the Postgres database password",
		Long:      "The rotation is both done in GCP and in the Kubernetes secret",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgrespasswordrotate.Before,
		Run:    postgrespasswordrotate.Action,
	}
	postgresPrepareCommand := &cli.Command{
		Use:   "prepare",
		Short: "Prepare your postgres instance for use with personal accounts",
		Long: `Prepare will prepare the postgres instance by connecting using the
application credentials and modify the permissions on the public schema.
All IAM users in your GCP project will be able to connect to the instance.

This operation is only required to run once for each postgresql instance.`,
		ArgsShort: "appname",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Use:   "all-privs",
				Short: "Gives all privileges to users",
			},
			contextFlag(),
			namespaceFlag(),
			&cli.StringFlag{
				Use:   "schema",
				Value: "public",
				Short: "Schema to grant access to",
			},
		},
		Before: postgresprepare.Before,
		Run:    postgresprepare.Action,
	}
	postgresProxyCommand := &cli.Command{
		Use:       "proxy",
		Short:     "Create a proxy to a Postgres instance",
		Long:      "Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Use:     "port",
				Aliases: []string{"p"},
				Value:   5432,
			},
			&cli.StringFlag{
				Use:     "host",
				Aliases: []string{"H"},
				Value:   "localhost",
			},
			&cli.BoolFlag{
				Use:     "verbose",
				Aliases: []string{"v"},
			},
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgresproxy.Before,
		Run:    postgresproxy.Action,
	}
	postgresPsqlCommand := &cli.Command{
		Use:       "psql",
		Short:     "Connect to the database using psql",
		Long:      "Create a shell to the postgres instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Use:     "verbose",
				Aliases: []string{"v"},
			},
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgrespsql.Before,
		Run:    postgrespsql.Action,
	}
	postgresRevokeCommand := &cli.Command{
		Use:   "revoke",
		Short: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
		Long: `Revoke will revoke the role 'cloudsqliamuser' access to the
tables in the postgres instance. This is done by connecting using the application
credentials and modify the permissions on the public schema.

This operation is only required to run once for each postgresql instance.`,
		ArgsShort: "appname",
		Flags: []cli.Flag{
			contextFlag(),
			namespaceFlag(),
			&cli.StringFlag{
				Use:   "schema",
				Value: "public",
				Short: "Schema to revoke access from",
			},
		},
		Before: postgresrevoke.Before,
		Run:    postgresrevoke.Action,
	}
	postgresUsersCommand := &cli.Command{
		Use:   "users",
		Short: "Administrate users in your Postgres instance",
		Long:  "Command used for listing and adding users to database",
	}
	postgresUsersAddCommand := &cli.Command{
		Use:       "add",
		Short:     "Add user to a Postgres database",
		Long:      "Will grant a user access to tables in public schema.",
		ArgsShort: "appname username password",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Use:     "privilege",
				Aliases: []string{"p"},
				Value:   "select",
			},
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgresusersadd.Before,
		Run:    postgresusersadd.Action,
	}
	postgresUsersListCommand := &cli.Command{
		Use:       "list",
		Short:     "List users in a Postgres database",
		ArgsShort: "appname",
		Flags: []cli.Flag{
			contextFlag(),
			namespaceFlag(),
		},
		Before: postgresuserslist.Before,
		Run:    postgresuserslist.Action,
	}

	return app.Execute()
}

func setDefaults(c *cli.Command) {
	c.HideHelpCommand = true

	for i := range c.Commands {
		setDefaults(c.Commands[i])
	}
}

func contextFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Use:         "context",
		Aliases:     []string{"c"},
		Short:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}

func copyFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Use:         "copy",
		Aliases:     []string{"cp"},
		Short:       "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected",
		DefaultText: "Attach to the current 'live' pod",
	}
}

func namespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Use:         "namespace",
		Aliases:     []string{"n"},
		Short:       "The kubernetes `NAMESPACE` to use",
		DefaultText: "The namespace from your current kubeconfig context",
	}
}

func noWaitFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Use:   "no-wait",
		Short: "Do not wait for the job to complete",
	}
}

func dryRunFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Use:   "dry-run",
		Short: "Perform a dry run of the migration setup, without actually starting the migration",
	}
}
