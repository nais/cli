package command

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func Valkey(parentFlags *alpha.Alpha) *naistrix.Command {
	flags := &flag.Valkey{Alpha: parentFlags}
	return &naistrix.Command{
		Name:        "valkey",
		Title:       "Manage Valkey instances.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			createValkey(flags),
			deleteValkey(flags),
			describeValkey(flags),
			listValkeys(flags),
			updateValkey(flags),
		},
	}
}
