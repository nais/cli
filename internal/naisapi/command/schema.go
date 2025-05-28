package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/command/flag"
	"github.com/nais/cli/internal/output"
)

func schema(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Schema{Api: parentFlags}
	return cli.NewCommand("schema", "Outputs the Nais API GraphQL schema to stdout.",
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			s, err := naisapi.PullSchema(ctx, flags)
			if err != nil {
				return err
			}

			out.Println(s)
			return nil
		}),
	)
}
