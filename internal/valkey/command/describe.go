package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func describeValkey(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Describe{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "describe",
		Title:       "Describe a Valkey instance.",
		Description: "This command describes a Valkey instance.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
