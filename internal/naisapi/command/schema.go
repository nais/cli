package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/naisapi"
	"github.com/nais/cli/v2/internal/naisapi/command/flag"
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
