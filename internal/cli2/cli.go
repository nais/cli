package cli2

import (
	"github.com/spf13/cobra"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func Run() error {
	app := &cobra.Command{
		Use:     "nais",
		Short:   "A Nais cli",
		Long:    "Nais platform utility cli, respects consoledonottrack.com",
		Version: version + "-" + commit,
	}

	loginCommand := &cobra.Command{
		Use:   "login",
		Short: "Login using Google Auth.",
		Long:  "This is a wrapper around gcloud auth login --update-adc.",
		// Run:   gcp.LoginCommand,
	}

	app.AddCommand(loginCommand)
	app.AddCommand(kubeconfigcmd())
	app.AddCommand(validatecmd())
	app.AddCommand(debugcmd())
	app.AddCommand(aivencmd())
	app.AddCommand(devicecmd())

	return app.Execute()
}

// 	postgresCommand := &cobra.Command{
// 		Use:    "postgres",
// 		Short:  "Command used for connecting to Postgres",
// 		Before: postgres.Before,
// 	}
// 	postgresEnableAuditCommand := &cli.Command{
// 		Use:       "enable-audit",
// 		Short:     "Enable audit extension in Postgres database",
// 		Long:      "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgresaudit.Before,
// 		Run:    postgresaudit.Action,
// 	}
// 	postgresGrandtCommand := &cli.Command{
// 		Use:       "grant",
// 		Short:     "Grant yourself access to a Postgres database",
// 		Long:      "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgresgrant.Before,
// 		Run:    postgresgrant.Action,
// 	}
// 	postgresMigrateCommand := &cli.Command{
// 		Use:    "migrate",
// 		Short:  "Command used for migrating to a new Postgres instance",
// 		Before: postgresmigrate.Before,
// 	}
// 	postgresMigrateSetupCommand := &cli.Command{
// 		Use:       "setup",
// 		Short:     "Make necessary setup for a new migration",
// 		UsageText: "nais postgres migrate setup APP_NAME TARGET_INSTANCE_NAME",
// 		Long:      "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
// 		Flags: []cli.Flag{
// 			namespaceFlag(),
// 			contextFlag(),
// 			dryRunFlag(),
// 			noWaitFlag(),
// 			&cli.StringFlag{
// 				Use:         "tier",
// 				Short:       "The `TIER` of the new instance",
// 				Category:    "Target instance configuration",
// 				Sources:     cli.EnvVars("TARGET_INSTANCE_TIER"),
// 				DefaultText: "Source instance value",
// 				Run:         postgresmigratesetup.TierFlagAction,
// 			},
// 			&cli.BoolFlag{
// 				Use:         "disk-autoresize",
// 				Short:       "Enable disk autoresize for the new instance",
// 				Category:    "Target instance configuration",
// 				Sources:     cli.EnvVars("TARGET_INSTANCE_DISK_AUTORESIZE"),
// 				DefaultText: "Source instance value",
// 			},
// 			&cli.IntFlag{
// 				Use:         "disk-size",
// 				Short:       "The `DISK_SIZE` of the new instance",
// 				Category:    "Target instance configuration",
// 				Sources:     cli.EnvVars("TARGET_INSTANCE_DISKSIZE"),
// 				DefaultText: "Source instance value",
// 			},
// 			&cli.StringFlag{
// 				Use:         "type",
// 				Short:       "The `TYPE` of the new instance",
// 				Category:    "Target instance configuration",
// 				Sources:     cli.EnvVars("TARGET_INSTANCE_TYPE"),
// 				DefaultText: "Source instance value",
// 				Run:         postgresmigratesetup.TypeFlagAction,
// 			},
// 		},
// 		Before: postgresmigratesetup.Before,
// 		Run:    postgresmigratesetup.Action,
// 	}
// 	postgresMigratePromoteCommand := &cli.Command{
// 		Use:       "promote",
// 		Short:     "Promote the migrated instance to the new primary instance",
// 		UsageText: "nais postgres migrate promote APP_NAME TARGET_INSTANCE_NAME",
// 		Long:      "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
// 		Flags: []cli.Flag{
// 			namespaceFlag(),
// 			contextFlag(),
// 			dryRunFlag(),
// 			noWaitFlag(),
// 		},
// 		Before: postgresmigratepromote.Before,
// 		Run:    postgresmigratepromote.Action,
// 	}
// 	postgresMigrateFinalizeCommand := &cli.Command{
// 		Use:       "finalize",
// 		Short:     "Finalize the migration",
// 		UsageText: "nais postgres migrate finalize APP_NAME TARGET_INSTANCE_NAME",
// 		Long:      "Finalize will remove the source instance and associated resources after a successful migration.",
// 		Flags: []cli.Flag{
// 			namespaceFlag(),
// 			contextFlag(),
// 			dryRunFlag(),
// 		},
// 		Before: postgresmigratefinalize.Before,
// 		Run:    postgresmigratefinalize.Action,
// 	}
// 	postgresMigrateRollbackCommand := &cli.Command{
// 		Use:       "rollback",
// 		Short:     "Roll back the migration",
// 		UsageText: "nais postgres migrate rollback APP_NAME TARGET_INSTANCE_NAME",
// 		Long:      "Rollback will roll back the migration, and restore the application to use the original instance.",
// 		Flags: []cli.Flag{
// 			namespaceFlag(),
// 			contextFlag(),
// 			dryRunFlag(),
// 		},
// 		Before: postgresmigraterollback.Before,
// 		Run:    postgresmigraterollback.Action,
// 	}
//
// 	postgresPasswordCommand := &cli.Command{
// 		Use:   "password",
// 		Short: "Administrate Postgres password",
// 	}
// 	postgresPasswordRotateCommand := &cli.Command{
// 		Use:       "rotate",
// 		Short:     "Rotate the Postgres database password",
// 		Long:      "The rotation is both done in GCP and in the Kubernetes secret",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgrespasswordrotate.Before,
// 		Run:    postgrespasswordrotate.Action,
// 	}
// 	postgresPrepareCommand := &cli.Command{
// 		Use:   "prepare",
// 		Short: "Prepare your postgres instance for use with personal accounts",
// 		Long: `Prepare will prepare the postgres instance by connecting using the
// application credentials and modify the permissions on the public schema.
// All IAM users in your GCP project will be able to connect to the instance.
//
// This operation is only required to run once for each postgresql instance.`,
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			&cli.BoolFlag{
// 				Use:   "all-privs",
// 				Short: "Gives all privileges to users",
// 			},
// 			contextFlag(),
// 			namespaceFlag(),
// 			&cli.StringFlag{
// 				Use:   "schema",
// 				Value: "public",
// 				Short: "Schema to grant access to",
// 			},
// 		},
// 		Before: postgresprepare.Before,
// 		Run:    postgresprepare.Action,
// 	}
// 	postgresProxyCommand := &cli.Command{
// 		Use:       "proxy",
// 		Short:     "Create a proxy to a Postgres instance",
// 		Long:      "Update IAM policies by giving your user the a timed sql.cloudsql.instanceUser role, then start a proxy to the instance.",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			&cli.UintFlag{
// 				Use:     "port",
// 				Aliases: []string{"p"},
// 				Value:   5432,
// 			},
// 			&cli.StringFlag{
// 				Use:     "host",
// 				Aliases: []string{"H"},
// 				Value:   "localhost",
// 			},
// 			&cli.BoolFlag{
// 				Use:     "verbose",
// 				Aliases: []string{"v"},
// 			},
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgresproxy.Before,
// 		Run:    postgresproxy.Action,
// 	}
// 	postgresPsqlCommand := &cli.Command{
// 		Use:       "psql",
// 		Short:     "Connect to the database using psql",
// 		Long:      "Create a shell to the postgres instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			&cli.BoolFlag{
// 				Use:     "verbose",
// 				Aliases: []string{"v"},
// 			},
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgrespsql.Before,
// 		Run:    postgrespsql.Action,
// 	}
// 	postgresRevokeCommand := &cli.Command{
// 		Use:   "revoke",
// 		Short: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
// 		Long: `Revoke will revoke the role 'cloudsqliamuser' access to the
// tables in the postgres instance. This is done by connecting using the application
// credentials and modify the permissions on the public schema.
//
// This operation is only required to run once for each postgresql instance.`,
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			contextFlag(),
// 			namespaceFlag(),
// 			&cli.StringFlag{
// 				Use:   "schema",
// 				Value: "public",
// 				Short: "Schema to revoke access from",
// 			},
// 		},
// 		Before: postgresrevoke.Before,
// 		Run:    postgresrevoke.Action,
// 	}
// 	postgresUsersCommand := &cli.Command{
// 		Use:   "users",
// 		Short: "Administrate users in your Postgres instance",
// 		Long:  "Command used for listing and adding users to database",
// 	}
// 	postgresUsersAddCommand := &cli.Command{
// 		Use:       "add",
// 		Short:     "Add user to a Postgres database",
// 		Long:      "Will grant a user access to tables in public schema.",
// 		ArgsShort: "appname username password",
// 		Flags: []cli.Flag{
// 			&cli.StringFlag{
// 				Use:     "privilege",
// 				Aliases: []string{"p"},
// 				Value:   "select",
// 			},
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgresusersadd.Before,
// 		Run:    postgresusersadd.Action,
// 	}
// 	postgresUsersListCommand := &cli.Command{
// 		Use:       "list",
// 		Short:     "List users in a Postgres database",
// 		ArgsShort: "appname",
// 		Flags: []cli.Flag{
// 			contextFlag(),
// 			namespaceFlag(),
// 		},
// 		Before: postgresuserslist.Before,
// 		Run:    postgresuserslist.Action,
// 	}
//
// 	return app.Execute()
// }
//
// func setDefaults(c *cli.Command) {
// 	c.HideHelpCommand = true
//
// 	for i := range c.Commands {
// 		setDefaults(c.Commands[i])
// 	}
// }
//
// func contextFlag() *cli.StringFlag {
// 	return &cli.StringFlag{
// 		Use:         "context",
// 		Aliases:     []string{"c"},
// 		Short:       "The kubeconfig `CONTEXT` to use",
// 		DefaultText: "The current context in your kubeconfig",
// 	}
// }
//
// func copyFlag() *cli.BoolFlag {
// 	return &cli.BoolFlag{
// 		Use:         "copy",
// 		Aliases:     []string{"cp"},
// 		Short:       "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected",
// 		DefaultText: "Attach to the current 'live' pod",
// 	}
// }
//
// func namespaceFlag() *cli.StringFlag {
// 	return &cli.StringFlag{
// 		Use:         "namespace",
// 		Aliases:     []string{"n"},
// 		Short:       "The kubernetes `NAMESPACE` to use",
// 		DefaultText: "The namespace from your current kubeconfig context",
// 	}
// }
//
// func noWaitFlag() *cli.BoolFlag {
// 	return &cli.BoolFlag{
// 		Use:   "no-wait",
// 		Short: "Do not wait for the job to complete",
// 	}
// }
//
// func dryRunFlag() *cli.BoolFlag {
// 	return &cli.BoolFlag{
// 		Use:   "dry-run",
// 		Short: "Perform a dry run of the migration setup, without actually starting the migration",
// 	}
// }
