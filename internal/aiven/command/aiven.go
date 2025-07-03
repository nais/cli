package command

import (
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Aiven(parentFlags *root.Flags) *naistrix.Command {
	aivenFlags := &flag.Aiven{Flags: parentFlags}
	return &naistrix.Command{
		Name:        "aiven",
		Title:       "Manage Aiven services.",
		StickyFlags: aivenFlags,
		SubCommands: []*naistrix.Command{
			create(aivenFlags),
			get(aivenFlags),
			tidy(aivenFlags),
		},
	}
}
