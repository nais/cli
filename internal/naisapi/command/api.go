package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
)

func Api(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Api{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "api",
		Title:       "Interact with Nais API.",
		Description: "Commands for interacting with the Nais API.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			proxyCommand(flags),
			schemaCommand(flags),
		},
	}
}
