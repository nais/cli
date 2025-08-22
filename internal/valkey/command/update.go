package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func updateValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Update{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "update",
		Title:       "Update a Valkey instance.",
		Description: "This command updates an existing Valkey instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
