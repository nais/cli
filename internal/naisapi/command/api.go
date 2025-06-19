package command

import (
	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/naisapi/command/flag"
)

func Api(parentFlags *flag.Alpha) *cli.Command {
	flags := &flag.Api{Alpha: parentFlags}
	return &cli.Command{
		Name:        "api",
		Title:       "Interact with Nais API.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			proxy(flags),
			schema(flags),
			team(flags),
			teams(flags),
			status(flags),
		},
	}
}
