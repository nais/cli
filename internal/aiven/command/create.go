package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/cli"
)

func create(parentFlags *flag.Aiven) *cli.Command {
	createFlags := &flag.Create{Aiven: parentFlags, Expire: 1}

	return &cli.Command{
		Name:  "create",
		Short: "Grant a user access to an Aiven service.",
		ValidateFunc: func(_ context.Context, _ []string) error {
			if createFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			return nil
		},
		StickyFlags: createFlags,
		SubCommands: []*cli.Command{
			createKafka(createFlags),
			createOpenSearch(createFlags),
		},
	}
}
