package logout

import (
	"context"

	"github.com/nais/cli/internal/auth/common"
	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type loginFlags struct {
	*flag.Auth
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais auth login --nais\"."`
}

func LogoutDeprecated(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Auth{GlobalFlags: parentFlags}
	return Logout(flags, common.Deprecated)
}

func Logout(parentFlags *flag.Auth, modifiers ...func(*naistrix.Command)) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	cmd := &naistrix.Command{
		Name:        "logout",
		Title:       "Log out and remove credentials.",
		Description: "Log out of the Nais platform and remove credentials from your local machine.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Logout(ctx, out)
			}

			return gcloud.Logout(ctx, out, flags.IsVerbose())
		},
	}
	for _, modifier := range modifiers {
		modifier(cmd)
	}
	return cmd
}
