package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/naistrix"
)

func tidy(_ *flag.Aiven) *naistrix.Command {
	return &naistrix.Command{
		Name:        "tidy",
		Title:       "Clean up /tmp/aiven-secret-* files made by the Nais CLI.",
		Description: "Caution - This command will delete all files in '/tmp' folder starting with 'aiven-secret-'.",
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			return aiven.TidyLocalSecrets(out)
		},
	}
}
