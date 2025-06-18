package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
)

func schema(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Schema{Api: parentFlags}
	return &cli.Command{
		Name:  "schema",
		Title: "Outputs the Nais API GraphQL schema to stdout.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			s, err := naisapi.PullSchema(ctx, flags)
			if err != nil {
				return err
			}

			out.Println(s)
			return nil
		},
	}
}
