package postgres

import (
	"database/sql"
	"log"
	"strings"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare [app-name] [flags]",
	Short: "Prepare your postgres instance for use with personal accounts",
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
