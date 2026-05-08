package command

import (
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Kafka(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Kafka{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:         "kafka",
		Aliases:      []string{"kafkas"},
		Title:        "Interact with Kafka topics.",
		Description:  "Commands for managing Kafka topics and credentials for your team.",
		StickyFlags:  flags,
		ValidateFunc: validation.RequireTeam(flags),
		SubCommands: []*naistrix.Command{
			credentials(flags),
			grantAccess(flags),
			list(flags),
		},
	}
}
