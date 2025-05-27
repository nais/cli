package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/naisapi/command/flag"
	naisapischema "github.com/nais/cli/internal/naisapi/schema"
	"github.com/nais/cli/internal/output"
)

func schema(parentFlags *flag.Api) *cli.Command {
	flags := &flag.Schema{Api: parentFlags}
	return cli.NewCommand("schema", "Outputs the Nais API GraphQL schema to stdout.", cli.WithRun(func(ctx context.Context, output output.Output, _ []string) error {
		schema, err := naisapischema.Pull(ctx, flags)
		if err != nil {
			return err
		}

		output.Println(schema)
		return nil
	}))
}
