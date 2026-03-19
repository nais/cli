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
		Name:        "job",
		Aliases:     []string{"jobs"},
		Title:       "Interact with jobs.",
		Description: "Commands for managing and inspecting your team's jobs, including listing, viewing activity and issues, triggering runs, and tailing logs.",
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(flags.Team)
		},
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			activity(flags),
			trigger(flags),
			issues(flags),
			log(flags),
			run(flags),
		},
	}
}
