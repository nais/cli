package command

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Job(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Job{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:    "job",
		Aliases: []string{"jobs"},
		Title:   "Interact with jobs.",
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			issues(flags),
			log(flags),
		},
	}
}
