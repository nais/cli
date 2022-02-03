package postgres

import (
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
		return "", err
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
		fmt.Println("User already has permanent access to database")
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
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
		return false, err
	}
	bindings := &policyBindings{}
	if err := json.Unmarshal(out, bindings); err != nil {
		return false, err
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
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return false, cmd.Run()
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
