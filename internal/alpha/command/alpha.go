package command

import (
	naisapi "github.com/nais/cli/internal/naisapi/command"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Alpha(parentFlags *root.Flags) *naistrix.Command {
	flags := &flag.Alpha{Flags: parentFlags}
	return &naistrix.Command{
		Name:        "alpha",
		Title:       "Alpha versions of Nais CLI commands.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			naisapi.Api(flags),
		},
	}
}
