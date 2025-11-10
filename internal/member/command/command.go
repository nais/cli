package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/naistrix"
)

func Members(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Member{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "members",
		Title:       "Interact with Nais team members.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			list(flags),
			add(flags),
			remove(flags),
			setRole(flags),
		},
	}
}
