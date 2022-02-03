package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var grantCmd = &cobra.Command{
	Use:   "grant [app-name] [flags]",
	Short: "Grant yourself access to a Postgres database.",
	Long: `Grant yourself access to a Postgres database.

	This is done by temporarily adding your user to the list of users that can administrate Cloud SQL instances and creating a user with your email.`,
	Args: cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		ctx := command.Context()

		dbInfo, err := NewDBInfo(appName, namespace, context)
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

		if err := grantUserAccess(ctx, projectID, "roles/cloudsql.admin", 5*time.Minute); err != nil {
			return err
		}
		if err := createSQLUser(ctx, projectID, connectionName); err != nil {
			fmt.Fprintln(os.Stderr, "Error creating SQL user. One might already exist.")
			return err
		}

		return nil
	},
}

func createSQLUser(ctx context.Context, projectID, instance string) error {
	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"sql",
		"users",
		"create",
		email,
		"--instance", strings.Split(instance, ":")[2],
		"--type", "cloud_iam_user",
		"--project", projectID,
	}

	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
