package logout

import (
	"context"

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
	flags := &loginFlags{Auth: &flag.Auth{GlobalFlags: parentFlags}}
	return &naistrix.Command{
		Name:        "logout",
		Title:       "Log out and remove credentials.",
		Description: "Log out of the Nais platform and remove credentials from your local machine.",
		Flags:       flags,
		Deprecated: naistrix.DeprecatedWithReplacementFunc(func(context.Context, *naistrix.Arguments) []string {
			cmd := []string{"auth", "logout"}
			if flags.Nais {
				cmd = append(cmd, "-n")
			}

			return cmd
		}),
	}
}

func Logout(parentFlags *flag.Auth) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	return &naistrix.Command{
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
}
