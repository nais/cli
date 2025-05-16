package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/k8s"
	postgrescmd "github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/audit"
	"github.com/nais/cli/internal/postgres/grant"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/finalize"
	"github.com/nais/cli/internal/postgres/migrate/promote"
	"github.com/nais/cli/internal/postgres/migrate/rollback"
	"github.com/nais/cli/internal/postgres/migrate/setup"
	"github.com/nais/cli/internal/postgres/password/rotate"
	"github.com/nais/cli/internal/postgres/prepare"
	"github.com/nais/cli/internal/postgres/proxy"
	"github.com/nais/cli/internal/postgres/psql"
	"github.com/nais/cli/internal/postgres/revoke"
	"github.com/nais/cli/internal/postgres/users/add"
	"github.com/nais/cli/internal/postgres/users/list"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func postgres(*root.Flags) *cobra.Command {
	cmdFlags := &postgrescmd.Flags{}
	cmd := &cobra.Command{
		Use:   "postgres",
		Short: "Manage SQL instances.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			_, err := gcp.ValidateAndGetUserLogin(cmd.Context(), false)
			return err
		},
	}

	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()

	cmd.PersistentFlags().StringVarP(&cmdFlags.Namespace, "namespace", "n", defaultNamespace, "The kubernetes `NAMESPACE` to use.")
	cmd.PersistentFlags().StringVarP(&cmdFlags.Context, "context", "c", defaultContext, "The kubeconfig `CONTEXT` to use.")

	migrateArguments := func(args []string) migrate.Arguments {
		return migrate.Arguments{
			ApplicationName:    args[0],
			TargetInstanceName: args[1],
		}
	}
	migrateCmdFlags := &migrate.Flags{Flags: cmdFlags}
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate to a new SQL instance.",
	}
	migrateCmd.PersistentFlags().BoolVar(&migrateCmdFlags.DryRun, "dry-run", false, "Perform a dry run.")

	migrateSetupCmdFlags := &setup.Flags{Flags: migrateCmdFlags}
	migrateSetupCmd := &cobra.Command{
		Use:   "setup APP_NAME TARGET_SQL_INSTANCE_NAME",
		Short: "Make necessary setup for a new SQL instance migration.",
		Long:  "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !strings.HasPrefix(cmd.Flag("tier").Value.String(), "db-") {
				return fmt.Errorf("tier must start with `db-`")
			}

			if !strings.HasPrefix(cmd.Flag("type").Value.String(), "POSTGRES_") {
				return fmt.Errorf("instance type must start with `POSTGRES_`")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.Run(
				cmd.Context(),
				migrateArguments(args),
				migrateSetupCmdFlags,
			)
		},
	}

	diskSize := 0
	if v, err := strconv.Atoi(os.Getenv("TARGET_INSTANCE_DISKSIZE")); err == nil {
		diskSize = v
	}

	diskAutoResize := false
	if v, err := strconv.ParseBool(os.Getenv("TARGET_INSTANCE_DISK_AUTORESIZE")); err == nil {
		diskAutoResize = v
	}

	migrateSetupCmd.Flags().BoolVar(&migrateSetupCmdFlags.NoWait, "no-wait", false, "Do not wait for the job to complete.")
	migrateSetupCmd.Flags().StringVar(&migrateSetupCmdFlags.Tier, "tier", os.Getenv("TARGET_INSTANCE_TIER"), "The `TIER` of the new instance.")
	migrateSetupCmd.Flags().BoolVar(&migrateSetupCmdFlags.DiskAutoResize, "disk-autoresize", diskAutoResize, "Enable disk autoresize for the new instance.")
	migrateSetupCmd.Flags().IntVar(&migrateSetupCmdFlags.DiskSize, "disk-size", diskSize, "The `DISK_SIZE` of the new instance.")
	migrateSetupCmd.Flags().StringVar(&migrateSetupCmdFlags.InstanceType, "type", os.Getenv("TARGET_INSTANCE_TYPE"), "The `TYPE` of the new instance.")

	_ = migrateSetupCmd.MarkFlagRequired("tier")
	_ = migrateSetupCmd.MarkFlagRequired("disk-autoresize")
	_ = migrateSetupCmd.MarkFlagRequired("disk-size")
	_ = migrateSetupCmd.MarkFlagRequired("type")

	migratePromoteCmdFlags := &promote.Flags{Flags: migrateCmdFlags}
	migratePromoteCmd := &cobra.Command{
		Use:   "promote APP_NAME TARGET_SQL_INSTANCE_NAME",
		Short: "Promote the migrated instance to the new primary instance.",
		Long:  "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return promote.Run(
				cmd.Context(),
				migrateArguments(args),
				migratePromoteCmdFlags,
			)
		},
	}
	migratePromoteCmd.Flags().BoolVar(&migratePromoteCmdFlags.NoWait, "no-wait", false, "Do not wait for the job to complete.")

	migrateFinalizeCmdFlags := &finalize.Flags{Flags: migrateCmdFlags}
	migrateFinalizeCmd := &cobra.Command{
		Use:   "finalize APP_NAME TARGET_SQL_INSTANCE_NAME",
		Short: "Finalize the migration.",
		Long:  "Finalize will remove the source instance and associated resources after a successful migration.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finalize.Run(
				cmd.Context(),
				migrateArguments(args),
				migrateFinalizeCmdFlags,
			)
		},
	}

	migrateRollbackCmdFlags := &rollback.Flags{Flags: migrateCmdFlags}
	migrateRollbackCmd := &cobra.Command{
		Use:   "rollback APP_NAME TARGET_SQL_INSTANCE_NAME",
		Short: "Roll back the migration.",
		Long:  "Rollback will roll back the migration, and restore the application to use the original instance.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rollback.Run(
				cmd.Context(),
				migrateArguments(args),
				migrateRollbackCmdFlags,
			)
		},
	}

	migrateCmd.AddCommand(
		migrateSetupCmd,
		migratePromoteCmd,
		migrateFinalizeCmd,
		migrateRollbackCmd,
	)

	passwordCmd := &cobra.Command{
		Use:   "password",
		Short: "Manage SQL instance password.",
	}

	passwordRotateCmd := &cobra.Command{
		Use:   "rotate APP_NAME",
		Short: "Rotate the SQL instance password.",
		Long:  "The rotation is both done in GCP and in the Kubernetes secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rotate.Run(cmd.Context(), args[0], cmdFlags)
		},
	}

	passwordCmd.AddCommand(passwordRotateCmd)

	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users in your SQL instance.",
		Long:  "Command used for listing and adding users to a SQL instance",
	}

	usersAddCmdFlags := &add.Flags{Flags: cmdFlags}
	usersAddCmd := &cobra.Command{
		Use:   "add APP_NAME USERNAME PASSWORD",
		Short: "Add a user to a SQL instance.",
		Long:  "Will grant a user access to tables in public schema.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return add.Run(
				cmd.Context(),
				add.Arguments{
					ApplicationName: args[0],
					Username:        args[1],
					Password:        args[2],
				},
				usersAddCmdFlags,
			)
		},
	}
	usersAddCmd.Flags().StringVarP(&usersAddCmdFlags.Privilege, "privilege", "P", "select", "The privilege to grant to the user.")

	usersListCmd := &cobra.Command{
		Use:   "list APP_NAME",
		Short: "List users in a SQL instance database.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return list.Run(cmd.Context(), args[0], cmdFlags)
		},
	}
	usersCmd.AddCommand(
		usersAddCmd,
		usersListCmd,
	)

	enableAuditCmd := &cobra.Command{
		Use:   "enable-audit APP_NAME",
		Short: "Enable audit extension in SQL instance database.",
		Long:  "This is done by creating pgaudit extension in the database and enabling audit logging for personal user accounts.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return audit.Run(cmd.Context(), args[0], cmdFlags)
		},
	}

	grantCmd := &cobra.Command{
		Use:   "grant APP_NAME",
		Short: "Grant yourself access to a SQL instance database.",
		Long:  "This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return grant.Run(cmd.Context(), args[0], cmdFlags)
		},
	}

	prepareCmdFlags := &prepare.Flags{Flags: cmdFlags}
	prepareCmd := &cobra.Command{
		Use:   "prepare APP_NAME",
		Short: "Prepare your SQL instance for use with personal accounts.",
		Long: `Prepare will prepare the SQL instance by connecting using the
		 application credentials and modify the permissions on the public schema.
		 All IAM users in your GCP project will be able to connect to the instance.
		
		 This operation is only required to run once for each SQL instance.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return prepare.Run(cmd.Context(), args[0], prepareCmdFlags)
		},
	}
	prepareCmd.Flags().BoolVar(&prepareCmdFlags.AllPrivileges, "all-privs", false, "Gives all privileges to users.")
	prepareCmd.Flags().StringVar(&prepareCmdFlags.Schema, "schema", "public", "Schema to grant access to.")

	proxyCmdFlags := &proxy.Flags{Flags: cmdFlags}
	proxyCmd := &cobra.Command{
		Use:   "proxy APP_NAME",
		Short: "Create a proxy to a SQL instance.",
		Long:  "Allows your user to connect to databases and starts a proxy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return proxy.Run(cmd.Context(), args[0], proxyCmdFlags)
		},
	}
	proxyCmd.Flags().UintVar(&proxyCmdFlags.Port, "port", 5432, "Port to use for the proxy.")
	proxyCmd.Flags().StringVar(&proxyCmdFlags.Host, "host", "localhost", "Host to use for the proxy.")

	psqlCmd := &cobra.Command{
		Use:   "psql APP_NAME",
		Short: "Connect to the database using psql.",
		Long:  "Create a shell to the SQL instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return psql.Run(cmd.Context(), args[0], cmdFlags)
		},
	}

	revokeCmdFlags := &revoke.Flags{Flags: cmdFlags}
	revokeCmd := &cobra.Command{
		Use:   "revoke APP_NAME",
		Short: "Revoke access to your SQL instance for the role 'cloudsqliamuser'.",
		Long: `Revoke will revoke the role 'cloudsqliamuser' access to the tables in the SQL instance.

This is done by connecting using the application credentials and modify the permissions on the public schema.

 This operation is only required to run once for each SQL instance.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return revoke.Run(cmd.Context(), args[0], revokeCmdFlags)
		},
	}
	revokeCmd.Flags().StringVar(&revokeCmdFlags.Schema, "schema", "public", "Schema to revoke access from.")

	cmd.AddCommand(
		migrateCmd,
		passwordCmd,
		usersCmd,
		enableAuditCmd,
		grantCmd,
		prepareCmd,
		proxyCmd,
		psqlCmd,
		revokeCmd,
	)

	return cmd
}
