package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/user/command/flag"
	"github.com/nais/naistrix"
)

func Users(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Users{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "users",
		Aliases:     []string{"user"},
		Title:       "View Nais users.",
		Description: "Commands for listing users on the Nais platform.",
		StickyFlags: f,
		SubCommands: []*naistrix.Command{
			list(f),
		},
	}
}
