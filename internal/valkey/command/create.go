package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func createValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Create{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "create",
		Title:       "Create a Valkey instance.",
		Description: "This command creates a Valkey instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
