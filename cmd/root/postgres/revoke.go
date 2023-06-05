package postgres

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const revokeHelp = `Revoke will revoke the role 'cloudsqliamuser' access to the
tables in the postgres instance. This is done by connecting using the application
credentials and modify the permissions on the public schema.

This operation is only required to run once for each postgresql instance.`

var revokeDdlStatements = []string{
	"alter default privileges in schema public revoke ALL on tables from cloudsqliamuser;",
	"alter default privileges in schema public revoke ALL on sequences from cloudsqliamuser;",
	"revoke ALL on all tables in schema public from cloudsqliamuser;",
	"revoke ALL on all sequences in schema public from cloudsqliamuser;",
}

var revokeCmd = &cobra.Command{
	Use:   "revoke [app-name] [flags]",
	Short: "Revoke access to your postgres instance for the role 'cloudsqliamuser'",
	Long:  revokeHelp,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		databaseName := viper.GetString(cmd.DatabaseFlag)
		dbInfo, err := NewDBInfo(appName, namespace, context, databaseName)
		if err != nil {
			return err
		}

		ctx := command.Context()

		fmt.Println(revokeHelp)

		fmt.Print("\nAre you sure you want to continue (y/N): ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
			return fmt.Errorf("cancelled by user")
		}

		connectionInfo, err := dbInfo.DBConnection(ctx)
		if err != nil {
			return err
		}

		db, err := sql.Open("cloudsqlpostgres", connectionInfo.ConnectionString())
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		for _, ddl := range revokeDdlStatements {
			_, err = db.ExecContext(ctx, ddl)
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	},
}
