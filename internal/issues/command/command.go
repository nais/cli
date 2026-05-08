package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Issues(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Issues{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:         "issues",
		Aliases:      []string{"issue"},
		Title:        "Manage issues.",
		Description:  "Commands for listing and managing critical issues detected for your team's workloads.",
		StickyFlags:  flags,
		ValidateFunc: validation.RequireTeam(flags),
		SubCommands: []*naistrix.Command{
			listIssues(flags),
		},
	}
}
