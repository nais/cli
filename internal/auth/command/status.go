package command

import (
	"context"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func status(parentFlags *flag.Auth) *naistrix.Command {
	f := &flag.Status{Auth: parentFlags, Output: "table"}
	return &naistrix.Command{
		Name:            "status",
		Title:           "Show authentication status for Nais and Google Cloud.",
		SkipPreRunHooks: []string{auth.NaisLoginHookName},
		Flags:           f,
		Description:     "Check both login types without starting a login flow. Nais uses the current API credentials, refreshing the session if necessary. Google Cloud checks the active gcloud account and the presence of Application Default Credentials; token validity is not verified. Missing credentials or an uninstalled gcloud are reported normally; other check failures return an error after both statuses are displayed. Use --output json for stable status values (authenticated, not_logged_in, not_installed, check_failed), null accounts when unavailable, and optional diagnostic details.",
		Examples: []naistrix.CommandExample{
			{Description: "Show the status of both login types without logging in."},
			{Description: "Get machine-readable status for both providers.", Command: "--output json"},
		},
		ValidateFlagsFunc: func(context.Context) error {
			if f.Output != "table" && f.Output != "json" {
				return naistrix.Errorf("unsupported output format %q: expected table or json", f.Output)
			}
			return nil
		},
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			statuses, statusErr := auth.LoginStatuses(ctx)
			if err := renderLoginStatuses(out, statuses, f.Output); err != nil {
				return err
			}
			return statusErr
		},
	}
}

func renderLoginStatuses(out *naistrix.OutputWriter, report auth.LoginStatusReport, format flag.StatusOutput) error {
	if format == "json" {
		return out.JSON(output.JSONWithPrettyOutput()).Render(report)
	}
	type row struct{ Provider, Status, Account, Details string }
	rows := make([]row, 0, 2)
	for _, status := range []auth.LoginStatus{report.Nais, report.Gcloud} {
		account, details := "-", "-"
		if status.Account != nil {
			account = *status.Account
		}
		if status.Details != "" {
			details = status.Details
		}
		rows = append(rows, row{
			Provider: status.Provider, Status: status.Status.String(), Account: account, Details: details,
		})
	}
	return out.Table().Render(rows)
}
