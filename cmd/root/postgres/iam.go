package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func currentEmail(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "account")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("currentEmail: unable to retrieve email: %w\n%v", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func grantUserAccess(ctx context.Context, projectID, role string, duration time.Duration) error {
	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	exists, err := cleanupPermissions(ctx, projectID, email, role, "nais_cli_access")
	if err != nil {
		return err
	}
	if exists {
		fmt.Println("User already has permanent access to database, will not grant temporary access")
		return nil
	}

	args := []string{
		"projects",
		"add-iam-policy-binding",
		projectID,
		"--member", "user:" + email,
		"--role", role,
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
		io.Copy(os.Stdout, buf)
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
	}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, string(e.Stderr))
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
	}
	cmd = exec.CommandContext(ctx, "gcloud", args...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		io.Copy(os.Stdout, buf)
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
