package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listIssues(parentFlags *flag.Issues) *naistrix.Command {
	flags := &flag.List{Issues: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List issues.",
		Description: "This command lists all issues for a given team.",
		Flags:       flags,
		Examples: []naistrix.Example{
			{
				Description: "List all issues for the team.",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			filters, err := issues.ParseFilter(flags)
			if err != nil {
				return fmt.Errorf("parse filter: %w", err)
			}
			ret, err := issues.GetAll(ctx, flags.Team, filters)
			if err != nil {
				return fmt.Errorf("fetching issues: %w", err)
			}

			if len(ret) == 0 {
				out.Infoln("No issues found")
				return nil
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				ID           string          `json:"id" hidden:"true"`
				Severity     issues.Severity `json:"severity"`
				Environment  string          `json:"environment"`
				ResourceName output.Link     `json:"resource_name" heading:"Resource Name"`
				ResourceType string          `json:"resource_type" heading:"Resource Type"`
				Message      string          `json:"message"`
			}

			entries := make([]entry, 0, len(ret))
			for _, i := range ret {
				entries = append(entries, entry{
					ID:          i.ID,
					Severity:    i.Severity,
					Environment: i.Environment,
					ResourceName: output.Link{
						Name: i.ResourceName,
						URL:  issueResourceURL(user.ConsoleHost(), flags.Team, i.Environment, i.ResourceType, i.ResourceName),
					},
					ResourceType: i.ResourceType,
					Message:      i.Message,
				})
			}

			return out.Table().Render(entries)
		},
	}
}

func issueResourceURL(host, team, environment, resourceType, resourceName string) string {
	switch resourceType {
	case "Application":
		return fmt.Sprintf("https://%s/team/%s/%s/app/%s", host, team, environment, resourceName)
	case "Job":
		return fmt.Sprintf("https://%s/team/%s/%s/job/%s", host, team, environment, resourceName)
	case "OpenSearch":
		return fmt.Sprintf("https://%s/team/%s/%s/opensearch/%s", host, team, environment, resourceName)
	case "SqlInstance":
		return fmt.Sprintf("https://%s/team/%s/%s/postgres/%s", host, team, environment, resourceName)
	case "Valkey":
		return fmt.Sprintf("https://%s/team/%s/%s/valkey/%s", host, team, environment, resourceName)
	case "Unleash", "UnleashInstance":
		return fmt.Sprintf("https://%s/team/%s/unleash", host, team)
	default:
		return ""
	}
}
