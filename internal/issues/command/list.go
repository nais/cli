package command

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

func listIssues(parentFlags *flag.Issues) *naistrix.Command {
	flags := &flag.List{Issues: parentFlags}
	return &naistrix.Command{
		Name:         "list",
		Title:        "List issues.",
		Description:  "This command lists all issues for a given team.",
		Flags:        flags,
		ValidateFunc: validation.TeamValidator(flags.Team),
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
			issues, err := issues.GetAll(ctx, flags.Team, filters)
			if err != nil {
				return fmt.Errorf("fetching issues: %w", err)
			}

			data := pterm.TableData{
				{
					"Issue",
					"Severity",
					"Resource Name",
					"Resource Type",
					"Environment",
					"Message",
				},
			}

			width, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				fmt.Println("could not get terminal size:", err)
				width = 160
			}

			for _, i := range issues {
				data = append(data, []string{
					i.IssueType,
					i.Severity,
					i.ResourceName,
					i.ResourceType,
					i.Environment,
					truncateString(i.Message, width-100),
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}

func truncateString(str string, max int) string {
	truncated := ""
	count := 0
	if len(str) < max {
		return str
	}

	for _, char := range str {
		truncated += string(char)
		count++
		if count >= max {
			break
		}
	}
	return truncated + "[...]"
}
