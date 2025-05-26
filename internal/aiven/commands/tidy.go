package commands

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
)

func tidy(_ *aivenFlags) *cli.Command {
	return cli.NewCommand("tidy", "Clean up /tmp/aiven-secret-* files made by the Nais CLI.",
		cli.WithLong(`Clean up /tmp/aiven-secret-* files made by the Nais CLI

Caution - This command will delete all files in "/tmp" folder starting with "aiven-secret-".`),
		cli.WithRun(func(ctx context.Context, _ output.Output, _ []string) error {
			return aiven.TidyLocalSecrets()
		}),
	)
}
