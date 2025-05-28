package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Login(_ *root.Flags) *cli.Command {
	cmdFlagNais := false
	return cli.NewCommand("login", "Log in to the Nais platform.",
		cli.InCommandGroup(cli.GroupAuthentication),
		cli.WithLong(`Log in to the Nais platform, uses "gcloud auth login --update-adc" by default.`),
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			if cmdFlagNais {
				return naisapi.Login(ctx, out)
			}

			return gcp.Login(ctx, out)
		}),
		cli.WithFlag("nais", "n", "Login using login.nais.io instead of gcloud.", &cmdFlagNais),
	)
}
