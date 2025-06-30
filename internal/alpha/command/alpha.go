package command

import (
	"github.com/nais/cli/pkg/cli/v2"
	naisapi "github.com/nais/cli/v2/internal/naisapi/command"
	"github.com/nais/cli/v2/internal/naisapi/command/flag"
	"github.com/nais/cli/v2/internal/root"
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
