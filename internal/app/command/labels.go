package command

import (
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func labels(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Labels{App: parentFlags}
	return &naistrix.Command{
		Name:        "labels",
		Title:       "Manage labels for an application.",
		Description: "Commands for listing and updating labels on an application.",
		Flags:       flags,
		SubCommands: []*naistrix.Command{
			labelsList(flags),
			labelsSet(flags),
		},
	}
}
