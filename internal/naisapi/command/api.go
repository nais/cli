package command

import (
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi/command/flag"
)

func Api(parentFlags *flag.Alpha) *cli.Command {
	flags := &flag.Api{Alpha: parentFlags}
	return cli.NewCommand("api", "Interact with Nais API.",
		cli.WithSubCommands(
			proxy(flags),
			schema(flags),
			teams(flags),
		),
	)
}
