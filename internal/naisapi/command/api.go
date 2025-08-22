package command

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
)

func Api(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Api{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "api",
		Title:       "Interact with Nais API.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			proxy(flags),
			schema(flags),
			team(flags),
			teams(flags),
			status(flags),
		},
	}
}
