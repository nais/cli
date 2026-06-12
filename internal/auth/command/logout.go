package command

import (
	"context"
	"os"

	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

func logout(parentFlags *flag.Auth) *naistrix.Command {
	flags := &flag.Logout{Auth: parentFlags}
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
