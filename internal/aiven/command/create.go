package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/naistrix"
)

func create(parentFlags *flag.Aiven) *naistrix.Command {
	createFlags := &flag.Create{Aiven: parentFlags, Expire: 1}

	return &naistrix.Command{
		Name:  "create",
		Title: "Grant a user access to an Aiven service.",
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			if createFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			return nil
		},
		StickyFlags: createFlags,
		SubCommands: []*naistrix.Command{
			createKafka(createFlags),
			createOpenSearch(createFlags),
		},
	}
}
