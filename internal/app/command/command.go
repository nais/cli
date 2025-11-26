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
		Name:  "app",
		Title: "Interact with applications.",
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			issues(flags),
			restart(flags),
		},
	}
}
