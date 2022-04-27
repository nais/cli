package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listUsersCmd = &cobra.Command{
	Use:   "list [app-name] [flags]",
	Short: "List users in a Postgres database.",
	Long:  `List users in a Postgres database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
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

		if err := listUsers(ctx, db); err != nil {
			return err
		}

		return nil
	},
}

func listUsers(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT usename FROM pg_catalog.pg_user;")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Users in database:")
	for rows.Next() {
		var d struct {
			User string `field:"usename"`
		}
		if err := rows.Scan(&d.User); err != nil {
			return err
		}

		fmt.Println(d.User)
	}

	return err
}
