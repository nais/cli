package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func listIssues(parentFlags *flag.Issues) *naistrix.Command {
	flags := &flag.List{Issues: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List issues.",
		Description: "This command lists all issues for a given team.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		ValidateFunc: func(_ context.Context, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("team cannot be empty")
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "List all issues for the team named my-team.",
				Command:     "my-team",
			},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			issues, err := issues.GetAll(ctx, args[0])
			if err != nil {
				return fmt.Errorf("fetching issues: %w", err)
			}

			data := pterm.TableData{
				{
					"Type",
					"Environment",
					"Severity",
					"Resource Name",
					"Resource Type",
					"Message",
				},
			}
			for _, i := range issues {
				data = append(data, []string{
					i.IssueType,
					i.Environment,
					i.Severity,
					i.ResourceName,
					i.ResourceType,
					i.Message,
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}
