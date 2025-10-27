package command

import (
	"context"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/naistrix"
)

func schemaCommand(parentFlags *flag.Api) *naistrix.Command {
	flags := &flag.Schema{Api: parentFlags}
	return &naistrix.Command{
		Name:  "schema",
		Title: "Outputs the Nais API GraphQL schema to stdout.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			s, err := naisapi.PullSchema(ctx, flags)
			if err != nil {
				return err
			}

			out.Println(s)
			return nil
		},
	}
}
