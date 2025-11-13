package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/naistrix"
)

func Issues(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Issues{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "issues",
		Title:       "Manage issues.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listIssues(flags),
		},
	}
}
