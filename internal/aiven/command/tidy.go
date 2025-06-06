package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/cli"
)

func tidy(_ *flag.Aiven) *cli.Command {
	return &cli.Command{
		Name:  "tidy",
		Short: "Clean up /tmp/aiven-secret-* files made by the Nais CLI.",
		Long: `Clean up /tmp/aiven-secret-* files made by the Nais CLI

Caution - This command will delete all files in "/tmp" folder starting with "aiven-secret-".`,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return aiven.TidyLocalSecrets(out)
		},
	}
}
