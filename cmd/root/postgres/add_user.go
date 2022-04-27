package postgres

import (
	"database/sql"
	"fmt"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addUserCmd = &cobra.Command{
	Use:   "add [username] [password] [app-name] [flags]",
	Short: "Add user to a Postgres database.",
	Long: `Add user to a Postgres database.

Grant user access to tables in public schema.`,
	Args: cobra.ExactArgs(3),
	RunE: func(command *cobra.Command, args []string) error {
		user := args[0]
		password := args[1]
		appName := args[2]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		privilege := viper.GetString(cmd.PrivilegeFlag)
		ctx := command.Context()

		dbInfo, err := NewDBInfo(appName, namespace, context)
		if err != nil {
			return err
		}

		connectionInfo, err := dbInfo.DBConnection(ctx)
		if err != nil {
			return err
		}

		db, err := sql.Open("cloudsqlpostgres", connectionInfo.ConnectionString())
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE USER %v WITH ENCRYPTED PASSWORD '%v' NOCREATEDB;", user, password))
		if err != nil {
			return err
		}
		fmt.Println("Created user", user)

		_, err = db.ExecContext(ctx, fmt.Sprintf("alter default privileges in schema public grant %v on tables to %q;", privilege, user))
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, fmt.Sprintf("grant %v on all tables in schema public to %q;", privilege, user))
		if err != nil {
			return err
		}

		return nil
	},
}
