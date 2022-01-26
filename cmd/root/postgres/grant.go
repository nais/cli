package postgres

import (
	"context"
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
	Short: "Grant yourself access to a Postgres database",
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

		projectID, err := dbInfo.ProjectID(ctx)
		if err != nil {
			return err
		}

		connectionName, err := dbInfo.ConnectionName(ctx)
		if err != nil {
			return err
		}

		if err := grantAccess(ctx, projectID, 1*time.Hour); err != nil {
			return err
		}
		if err := createSQLUser(ctx, projectID, connectionName); err != nil {
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

func grantAccess(ctx context.Context, projectID string, duration time.Duration) error {
	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"projects",
		"add-iam-policy-binding",
		projectID,
		"--member", "user:" + email,
		"--role", "roles/cloudsql.admin",
	}

	if duration > 0 {
		timestamp := time.Now().Add(duration).UTC().Format(time.RFC3339)
		args = append(args,
			"--condition",
			"expression=request.time < timestamp('"+timestamp+"'),title=temp_access",
		)
	}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func currentEmail(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
