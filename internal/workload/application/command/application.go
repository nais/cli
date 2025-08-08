package command

import (
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/internal/workload/application/command/flag"
	"github.com/nais/naistrix"
)

func Application(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Application{Flags: parentFlags}

	return &naistrix.Command{
		Name:        "application",
		Aliases:     []string{"app"},
		Title:       "Manage applications.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			createCommand(flags),
		},
	}
}
