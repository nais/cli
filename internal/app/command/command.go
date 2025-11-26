package command

import (
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func App(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.App{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "app",
		Title:       "Interact with applications.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			issues(flags),
			restart(flags),
		},
	}
}
