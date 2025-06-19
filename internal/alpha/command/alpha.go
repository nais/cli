package command

import (
	naisapi "github.com/nais/cli/internal/naisapi/command"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func Alpha(parentFlags *root.Flags) *cli.Command {
	flags := &flag.Alpha{Flags: parentFlags}
	return &cli.Command{
		Name:        "alpha",
		Title:       "Alpha versions of Nais CLI commands.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			naisapi.Api(flags),
		},
	}
}
