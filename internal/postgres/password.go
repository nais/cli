package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/liberator/pkg/keygen"
	"github.com/nais/naistrix"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RotatePassword(ctx context.Context, appName string, fl *flag.Password, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	if _, err := GetSecretValues(ctx, appName, fl.Postgres, ReasonPasswordRotate, out); err != nil {
		return err
	}

	dbInfo, err := NewDBInfo(ctx, appName, fl.Namespace, fl.Context)
	if err != nil {
		return err
	}

	cloudSQLDBInfo, err := dbInfo.ToCloudSQLDBInfo()
	if err != nil {
		return err
	}

	projectID, err := cloudSQLDBInfo.ProjectID(ctx)
	if err != nil {
		return err
	}

	dbConnectionInfo, err := cloudSQLDBInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	out.Println("Grant user cloudsql.admin access for 5 minutes")
	err = grantUserAccess(ctx, projectID, "roles/cloudsql.admin", 5*time.Minute, out)
	if err != nil {
		return err
	}

	out.Println("Generating new password")
	newPassword, err := generatePassword()
	if err != nil {
		return err
	}

	dbConnectionInfo.SetPassword(newPassword)

	out.Printf("Rotating password for user %v in database %v\n", dbConnectionInfo.username, dbConnectionInfo.dbName)
	err = rotatePasswordForDatabaseUser(ctx, projectID, dbConnectionInfo.instance, dbConnectionInfo.username, dbConnectionInfo.password)
	if err != nil {
		return err
	}

	out.Printf("Updating password in k8s secret google-sql-%v\n", cloudSQLDBInfo.appName)
	err = updateKubernetesSecret(ctx, cloudSQLDBInfo, dbConnectionInfo)
	if err != nil {
		return err
	}

	out.Println("Password rotated")
	return nil
}

func updateKubernetesSecret(ctx context.Context, dbInfo *CloudSQLDBInfo, dbConnectionInfo *ConnectionInfo) error {
	secret, err := dbInfo.k8sClient.CoreV1().Secrets(string(dbInfo.namespace)).Get(ctx, "google-sql-"+dbInfo.appName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to the k8s secret %q in %q: %w", "google-sql-"+dbInfo.appName, dbInfo.namespace, err)
	}

	jdbcUrlSet := false
	prefix := ""
	for key := range secret.Data {
		if strings.HasSuffix(key, "_PASSWORD") {
			secret.Data[key] = []byte(dbConnectionInfo.password)
		}
		if strings.HasSuffix(key, "_URL") {
			if strings.HasSuffix(key, "_JDBC_URL") && dbConnectionInfo.jdbcUrl != nil {
				secret.Data[key] = []byte(dbConnectionInfo.jdbcUrl.String())
				jdbcUrlSet = true
			} else if dbConnectionInfo.url != nil {
				secret.Data[key] = []byte(dbConnectionInfo.url.String())
				prefix = strings.TrimSuffix(key, "_URL")
			}
		}
	}

	if !jdbcUrlSet && dbConnectionInfo.jdbcUrl != nil && len(prefix) > 0 {
		key := prefix + "_JDBC_URL"
		secret.Data[key] = []byte(dbConnectionInfo.jdbcUrl.String())
	}

	_, err = dbInfo.k8sClient.CoreV1().Secrets(string(dbInfo.namespace)).Update(ctx, secret, v1.UpdateOptions{})
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
	err := cmd.Run()
	if err != nil {
		_, _ = io.Copy(os.Stdout, buf)
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
