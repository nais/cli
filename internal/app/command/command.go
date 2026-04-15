package command

import (
	"context"

	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func App(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.App{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "app",
		Aliases:     []string{"apps", "application", "applications"},
		Title:       "Interact with applications.",
		Description: "Commands for managing and inspecting your team's applications, including listing, viewing activity and issues, restarting, and tailing logs.",
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			activity(flags),
			issues(flags),
			restart(flags),
			log(flags),
			status(flags),
			env(flags),
			files(flags),
		},
	}
}
