package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Logout(_ *root.Flags) *cli.Command {
	cmdFlagNais := false
	return cli.NewCommand("logout", "Log out and remove credentials.",
		cli.InGroup(cli.GroupAuthentication),
		cli.WithLongDescription("Log out of the Nais platform and remove credentials from your local machine."),
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			if cmdFlagNais {
				return naisapi.Logout(ctx, out)
			}

			return gcp.Logout(ctx, out)
		}),
		cli.WithFlag("nais", "n", "Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\".", &cmdFlagNais),
	)
}
