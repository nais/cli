package command

import (
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Aiven(parentFlags *flags.GlobalFlags) *naistrix.Command {
	aivenFlags := &flag.Aiven{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "aiven",
		Title:       "Manage Aiven services.",
		StickyFlags: aivenFlags,
		SubCommands: []*naistrix.Command{
			create(aivenFlags),
			get(aivenFlags),
			tidy(aivenFlags),
			grantAccess(aivenFlags),
		},
	}
}
