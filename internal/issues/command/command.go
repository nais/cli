package command

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/naistrix"
)

func Issues(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Issues{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "issues",
		Title:       "Manage issues.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listIssues(flags),
		},
	}
}
