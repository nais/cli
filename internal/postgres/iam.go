package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/k8s"
)

func GrantAndCreateSQLUser(ctx context.Context, appName string, cluster k8s.Context, namespace string, out cli.Output) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
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

	out.Println("Grant user access")
	err = grantUserAccess(ctx, projectID, "roles/cloudsql.admin", 5*time.Minute, out)
	if err != nil {
		return err
	}

	out.Println("Create sql user")
	err = createSQLUser(ctx, projectID, connectionName)
	if err != nil {
		return fmt.Errorf("error creating SQL user. One might already exist: %v", err)
	}

	return nil
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

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_, _ = io.Copy(os.Stdout, buf)
		return fmt.Errorf("error running gcloud command: %w", err)
	}
	return nil
}

func currentEmail(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("currentEmail: unable to retrieve email: %w\n%v", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func grantUserAccess(ctx context.Context, projectID, role string, duration time.Duration, out cli.Output) error {
	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	exists, err := cleanupPermissions(ctx, projectID, email, role, "nais_cli_access")
	if err != nil {
		return err
	}

	if exists {
		out.Println("User already has permanent access to database, will not grant temporary access")
		return nil
	}

	args := []string{
		"projects",
		"add-iam-policy-binding",
		projectID,
		"--member", "user:" + email,
		"--role", role,
		"--billing-project", projectID,
	}

	if duration > 0 {
		timestamp := time.Now().Add(duration).UTC().Format(time.RFC3339)
		args = append(args,
			"--condition",
			formatCondition("request.time < timestamp('"+timestamp+"')", "nais_cli_access"),
		)
	}

	cmd := exec.CommandContext(ctx, "gcloud", args...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_, _ = io.Copy(os.Stdout, buf)
		return fmt.Errorf("grantUserAccess: error running gcloud command: %w", err)
	}
	return nil
}

func cleanupPermissions(ctx context.Context, projectID, email, role, conditionName string) (exists bool, err error) {
	args := []string{
		"projects",
		"get-iam-policy",
		projectID,
		"--format", "json",
		"--billing-project", projectID,
	}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			_, _ = fmt.Fprintln(os.Stderr, string(e.Stderr))
		}
		return false, fmt.Errorf("cleanupPermissions: error getting permissions: %w", err)
	}
	bindings := &policyBindings{}
	if err := json.Unmarshal(out, bindings); err != nil {
		return false, fmt.Errorf("cleanupPermissions: error unmarshaling json: %w", err)
	}

	expr := ""
OUTER:
	for _, binding := range bindings.Bindings {
		if binding.Role == role {
			for _, member := range binding.Members {
				if member == "user:"+email {
					if binding.Condition == nil {
						return true, nil
					}
					if binding.Condition.Title == conditionName {
						expr = formatCondition(binding.Condition.Expression, binding.Condition.Title)
						break OUTER
					}
				}
			}
		}
	}

	if expr == "" {
		return false, nil
	}

	args = []string{
		"projects",
		"remove-iam-policy-binding",
		projectID,
		"--member", "user:" + email,
		"--role", role,
		"--condition", expr,
		"--billing-project", projectID,
	}
	cmd = exec.CommandContext(ctx, "gcloud", args...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_, _ = io.Copy(os.Stdout, buf)
		return false, fmt.Errorf("cleanupPermissions: error running gcloud command: %w", err)
	}
	return false, nil
}

type policyBindings struct {
	Bindings []struct {
		Role      string   `json:"role"`
		Members   []string `json:"members"`
		Condition *struct {
			Title      string `json:"title"`
			Expression string `json:"expression"`
		} `json:"condition"`
	} `json:"bindings"`
}

func formatCondition(expr, title string) string {
	return fmt.Sprintf("expression=%v,title=%v", expr, title)
}

func ListUsers(ctx context.Context, appName string, cluster k8s.Context, namespace string, out cli.Output) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}

	rows, err := db.QueryContext(ctx, "SELECT usename FROM pg_catalog.pg_user;")
	if err != nil {
		return formatInvalidGrantError(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	out.Println("Users in database:")
	for rows.Next() {
		var d struct {
			User string `field:"usename"`
		}
		if err := rows.Scan(&d.User); err != nil {
			return err
		}

		out.Println(d.User)
	}

	return err
}

func AddUser(ctx context.Context, appName, username, password string, cluster k8s.Context, namespace, privilege string, out cli.Output) error {
	err := validateSQLVariables(username, password, privilege)
	if err != nil {
		return err
	}

	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE USER %v WITH ENCRYPTED PASSWORD '%v' NOCREATEDB;", username, password))
	if err != nil {
		return formatInvalidGrantError(err)
	}
	out.Printf("Created user: %v", username)

	_, err = db.ExecContext(ctx, fmt.Sprintf("alter default privileges in schema public grant %v on tables to %q;", privilege, username))
	if err != nil {
		return formatInvalidGrantError(err)
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("grant %v on all tables in schema public to %q;", privilege, username))
	if err != nil {
		return formatInvalidGrantError(err)
	}

	return nil
}

func validateSQLVariables(variables ...string) error {
	r, err := regexp.Compile("^([A-Za-z0-9-_]+)$")
	if err != nil {
		return err
	}

	for _, v := range variables {
		if match := r.MatchString(v); !match {
			return fmt.Errorf("invalid sql argument: %v (only letters, numbers, - and _ are allowed)", v)
		}
	}

	return nil
}
