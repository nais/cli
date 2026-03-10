package command

import (
	"context"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func listCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.List{Postgres: parentFlags}

	return &naistrix.Command{
		Name:  "list",
		Title: "List postgres instances for a team.",
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			out.Println("Not yet implemented.")
			return nil
		},
	}
}
