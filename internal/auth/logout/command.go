package logout

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Command(_ *root.Flags) *cli.Command {
	cmdFlagNais := false
	return cli.NewCommand("logout", "Log out and remove credentials.",
		cli.WithLong("Log out of the Nais platform and remove credentials from your local machine."),
		cli.WithRun(func(ctx context.Context, w output.Output, _ []string) error {
			if cmdFlagNais {
				return naisapi.Logout(ctx, w)
			}

			return gcp.Logout(ctx, w)
		}),
		cli.WithFlag("nais", "n", "Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\".", &cmdFlagNais),
	)
}
