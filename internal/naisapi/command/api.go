package command

import (
	"github.com/nais/cli/v2/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
)

func Api(parentFlags *flag.Alpha) *naistrix.Command {
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
