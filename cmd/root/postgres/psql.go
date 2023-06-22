package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var psqlCmd = &cobra.Command{
	Use:   "psql [app-name] [flags]",
	Short: "Connect to the database using psql",
	Long:  `Create a shell to the postgres instance by opening a proxy on a random port (see the proxy command for more info) and opening a psql shell.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		k8sContext := viper.GetString(cmd.ContextFlag)
		verbose := viper.GetBool(cmd.VerboseFlag)
		databaseName := viper.GetString(cmd.DatabaseFlag)

		psqlPath, err := exec.LookPath("psql")
		if err != nil {
			return err
		}

		dbInfo, err := NewDBInfo(appName, namespace, k8sContext, databaseName)
		if err != nil {
			return err
		}

		projectID, err := dbInfo.ProjectID(ctx)
		if err != nil {
			return err
		}

		connectionName, err := dbInfo.ConnectionName(ctx)
		if err != nil {
			return err
		}
		connectionInfo, err := dbInfo.DBConnection(ctx)
		if err != nil {
			return err
		}

		email, err := currentEmail(ctx)
		if err != nil {
			return err
		}

		portCh := make(chan int, 1)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			if !verbose {
				logging.DisableLogging()
			}

			if err := runProxy(ctx, projectID, connectionName, "localhost:0", portCh); err != nil {
				log.Println("ERROR:", err)
				cancel()
			}
		}()
		port := <-portCh

		fmt.Printf("Running proxy on localhost:%v\n", port)

		arguments := []string{
			"--host", "localhost",
			"--port", fmt.Sprintf("%d", port),
			"--username", email,
			"--dbname", connectionInfo.dbName,
		}

		cmd := exec.CommandContext(ctx, psqlPath, arguments...)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Env = os.Environ()

		return cmd.Run()
	},
}
