package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func deleteValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Delete{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete a Valkey instance.",
		Description: "This command deletes an existing Valkey instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
