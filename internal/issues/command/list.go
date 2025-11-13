package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/validation"
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
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
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

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(issues)
			}

			return out.Table().Render(issues)
		},
	}
}
