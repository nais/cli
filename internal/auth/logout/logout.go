package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

type flags struct {
	*root.Flags
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
}

func Logout(rootFlags *root.Flags) *cli.Command {
	flags := &flags{Flags: rootFlags}
	return &cli.Command{
		Name:  "logout",
		Short: "Log out and remove credentials.",
		Long:  "Log out of the Nais platform and remove credentials from your local machine.",
		Group: cli.GroupAuthentication,
		Flags: flags,
		RunFunc: func(ctx context.Context, out output.Output, _ []string) error {
			if flags.Nais {
				return naisapi.Logout(ctx, out)
			}

			return gcp.Logout(ctx, out)
		},
	}
}
