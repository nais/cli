package command

import (
	"context"

	"github.com/nais/cli/internal/activity/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Activity(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Activity{GlobalFlags: parentFlags}

	return &naistrix.Command{
		Name:        "activity",
		Title:       "List team activity.",
		Description: "View recent activity across all resources in a team, such as deployments, configuration changes, and other events.",
		StickyFlags: f,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return validation.CheckTeam(f.Team)
		},
		SubCommands: []*naistrix.Command{
			list(f),
		},
	}
}
