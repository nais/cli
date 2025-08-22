package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func listValkeys(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.List{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List existing Valkey instances.",
		Description: "This command lists all Valkey instances across your teams.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			// TODO: filter by team
			// FIXME
			return fmt.Errorf("not implemented yet")
		},
	}
}
