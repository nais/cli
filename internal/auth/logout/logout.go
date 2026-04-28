package logout

import (
	"context"
	"os"

	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

type loginFlags struct {
	*flag.Auth
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
	Yes  bool `name:"yes" short:"y" usage:"Automatically answer yes to all prompts."`
}

func Logout(parentFlags *flag.Auth) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	return &naistrix.Command{
		Name:            "logout",
		TopLevelAliases: []string{"logout"},
		Title:           "Log out and remove credentials.",
		Description:     "Log out of the Nais platform and remove credentials from your local machine.",
		Flags:           flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Logout(ctx, out)
			}

			if err := gcloud.Logout(ctx, out, flags.IsVerbose()); err != nil {
				return err
			}

			if term.IsTerminal(int(os.Stdin.Fd())) { // #nosec G115
				out.Println()
				if flags.Yes {
					return naisapi.Logout(ctx, out)
				}

				if result, err := input.Confirm("Would you like to also log out of Nais?", input.ConfirmWithDefaultTrue()); err != nil {
					return err
				} else if result {
					return naisapi.Logout(ctx, out)
				}
			}

			return nil
		},
	}
}
