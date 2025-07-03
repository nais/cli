package command

import (
	"context"

	"github.com/nais/cli/v2/internal/auth"
	"github.com/nais/cli/v2/internal/gcloud"
	"github.com/nais/cli/v2/internal/naisapi"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

type flags struct {
	*root.Flags
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
}

func Logout(rootFlags *root.Flags) *naistrix.Command {
	flags := &flags{Flags: rootFlags}
	return &naistrix.Command{
		Name:        "logout",
		Title:       "Log out and remove credentials.",
		Description: "Log out of the Nais platform and remove credentials from your local machine.",
		Group:       auth.GroupName,
		Flags:       flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			if flags.Nais {
				return naisapi.Logout(ctx, out)
			}

			return gcloud.Logout(ctx, out)
		},
	}
}
