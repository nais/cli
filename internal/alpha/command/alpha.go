package command

import (
	"github.com/nais/cli/internal/cli"
	naisapi "github.com/nais/cli/internal/naisapi/command"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/root"
)

func Alpha(parentFlags *root.Flags) *cli.Command {
	flags := &flag.Alpha{Flags: parentFlags}
	return &cli.Command{
		Name:        "alpha",
		Short:       "Alpha versions of Nais CLI commands.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			naisapi.Api(flags),
		},
	}
}
