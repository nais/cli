package postgres

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)

const prepareHelp = `Prepare will prepare the postgres instance by connecting using the
application credentials and modify the permissions on the public schema.
All IAM users in your GCP project will be able to connect to the instance.

This operation is only required to run once for each postgresql instance.`

var prepareCmd = &cobra.Command{
	Use:   "prepare [app-name] [flags]",
	Short: "Prepare your postgres instance for use with personal accounts",
	Long:  prepareHelp,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		dbInfo, err := NewDBInfo(appName, namespace, context)
		if err != nil {
			return err
		}

		ctx := command.Context()

		fmt.Println(prepareHelp)

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

		_, err = db.ExecContext(ctx, "grant all on all tables in schema public to cloudsqliamuser;")
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.ExecContext(ctx, "grant all on all sequences in schema public to cloudsqliamuser;")
		return err
	},
}

func getSecretDataValue(secret *corev1.Secret, suffix string) string {
	for name, val := range secret.Data {
		if strings.HasSuffix(name, suffix) {
			return string(val)
		}
	}
	return ""
}
