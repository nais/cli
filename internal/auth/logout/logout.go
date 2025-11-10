package command

import (
	"context"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type LogoutFlags struct {
	*flags.GlobalFlags
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
}

func Logout(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &LogoutFlags{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "logout",
		Title:       "Log out and remove credentials.",
		Description: "Log out of the Nais platform and remove credentials from your local machine.",
		Group:       auth.GroupName,
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Logout(ctx, out)
			}

			return gcloud.Logout(ctx, out, flags.IsVerbose())
		},
	}
}
