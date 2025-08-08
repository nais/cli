package command

import (
	"context"

	"github.com/nais/cli/internal/workload/application/command/flag"
	"github.com/nais/cli/internal/workload/application/create"
	"github.com/nais/naistrix"
)

func createCommand(parentFlags *flag.Application) *naistrix.Command {
	flags := &flag.Create{Application: parentFlags}
	return &naistrix.Command{
		Name:  "create",
		Title: "Create a new application.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return create.Run(flags, out)
		},
	}
}
