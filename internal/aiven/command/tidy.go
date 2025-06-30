package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/aiven"
	"github.com/nais/cli/v2/internal/aiven/command/flag"
)

func tidy(_ *flag.Aiven) *cli.Command {
	return &cli.Command{
		Name:        "tidy",
		Title:       "Clean up /tmp/aiven-secret-* files made by the Nais CLI.",
		Description: "Caution - This command will delete all files in '/tmp' folder starting with 'aiven-secret-'.",
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			return aiven.TidyLocalSecrets(out)
		},
	}
}
