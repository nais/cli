package command

import (
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/root"
)

func Aiven(parentFlags *root.Flags) *cli.Command {
	aivenFlags := &flag.Aiven{Flags: parentFlags}
	return &cli.Command{
		Name:        "aiven",
		Title:       "Manage Aiven services.",
		StickyFlags: aivenFlags,
		SubCommands: []*cli.Command{
			create(aivenFlags),
			get(aivenFlags),
			tidy(aivenFlags),
		},
	}
}
