package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nais/cli/cmd"
	"github.com/nais/liberator/pkg/keygen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var passwordRotateCmd = &cobra.Command{
	Use:   "rotate [app-name]",
	Short: "Rotate the Postgres database password.",
	Long:  `Rotate the Postgres database password, both in GCP and in the Kubernetes secret.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		databaseName := viper.GetString(cmd.DatabaseFlag)
		ctx := command.Context()

		dbInfo, err := NewDBInfo(appName, namespace, context, databaseName)
		if err != nil {
			return err
		}

		projectID, err := dbInfo.ProjectID(ctx)
		if err != nil {
			return err
		}

		dbConnectionInfo, err := dbInfo.DBConnection(ctx)
		if err != nil {
			return err
		}

		fmt.Println("Grant user cloudsql.admin access for 5 minutes")
		err = grantUserAccess(ctx, projectID, "roles/cloudsql.admin", 5*time.Minute)
		if err != nil {
			return err
		}

		fmt.Println("Generating new password")
		newPassword, err := generatePassword()
		if err != nil {
			return err
		}
		dbConnectionInfo.SetPassword(newPassword)

		fmt.Printf("Rotating password for user %v in database %v\n", dbConnectionInfo.username, dbConnectionInfo.dbName)
		err = rotatePasswordForDatabaseUser(ctx, projectID, dbConnectionInfo.instance, dbConnectionInfo.username, dbConnectionInfo.password)
		if err != nil {
			return err
		}

		fmt.Printf("Updating password in k8s secret google-sql-%v\n", dbInfo.appName)
		return updateKubernetesSecret(ctx, dbInfo, dbConnectionInfo)

		fmt.Println("Password rotated")
	},
}

func updateKubernetesSecret(ctx context.Context, dbInfo *DBInfo, dbConnectionInfo *ConnectionInfo) error {
	secret, err := dbInfo.k8sClient.CoreV1().Secrets(dbInfo.namespace).Get(ctx, "google-sql-"+dbInfo.appName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to the k8s secret %q in %q: %w", "google-sql-"+dbInfo.appName, dbInfo.namespace, err)
	}

	for key, _ := range secret.Data {
		if strings.HasSuffix(key, "_PASSWORD") {
			secret.Data[key] = []byte(dbConnectionInfo.password)
		}
		if strings.HasSuffix(key, "_URL") {
			secret.Data[key] = []byte(dbConnectionInfo.JDBCURL())
		}
	}

	_, err = dbInfo.k8sClient.CoreV1().Secrets(dbInfo.namespace).Update(ctx, secret, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed updating k8s secret %q in %q with new password: %w", "google-sql-"+dbInfo.appName, dbInfo.namespace, err)
	}

	return nil
}

func rotatePasswordForDatabaseUser(ctx context.Context, projectID, instance, username, password string) error {
	args := []string{
		"sql",
		"users",
		"set-password",
		username,
		"--password", password,
		"--instance", strings.Split(instance, ":")[2],
		"--project", projectID,
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		io.Copy(os.Stdout, buf)
		return fmt.Errorf("error running gcloud command: %w", err)
	}
	return nil
}

func generatePassword() (string, error) {
	key, err := keygen.Keygen(32)
	if err != nil {
		return "", fmt.Errorf("unable to generate secret for sql user: %s", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(key), nil
}
