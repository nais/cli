package command

import (
	"context"
	"errors"
	"os"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

func login(parentFlags *flag.Auth) *naistrix.Command {
	flags := &flag.Login{Auth: parentFlags}
	return &naistrix.Command{
		SkipPreRunHooks: []string{auth.NaisLoginHookName},
		Name:            "login",
		Title:           "Log in to the Nais platform.",
		TopLevelAliases: []string{"login"},
		Examples: []naistrix.CommandExample{
			{
				Description: "Log in to the Nais platform using gcloud.",
			},
			{
				Description: "Log in to the Nais platform using login.nais.io.",
				Command:     "-n",
			},
			{
				Description: "Log in to the Nais platform using gcloud and login.nais.io without prompts.",
				Command:     "-y",
			},
		},
		Description: `Uses "gcloud auth login --update-adc" by default.`,
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			if err := gcloud.Login(ctx, out, flags.IsVerbose()); err != nil {
				return err
			}

			if term.IsTerminal(int(os.Stdin.Fd())) { // #nosec G115
				return offerNaisLogin(ctx, out, flags.Yes, flags.ConfirmLogin, naisapi.GetAuthenticatedUser, naisapi.Login, input.Confirm)
			}

			return nil
		},
	}
}

func offerNaisLogin(ctx context.Context, out *naistrix.OutputWriter, autoAccept, confirmLogin bool, getUser func(context.Context) (naisapi.AuthenticatedUser, error), login naistrix.PreRunHookFunc, confirm func(string, ...input.ConfirmOptionFunc) (bool, error)) error {
	if _, err := getUser(ctx); !errors.Is(err, naisapi.ErrNeedsLogin) {
		return err
	}

	out.Println()
	out.Println("Many Nais commands require you to be logged in to both Google and Nais.")
	if autoAccept || !confirmLogin {
		return login(ctx, out)
	}

	result, err := confirm("Would you like to also log in to Nais?", input.ConfirmWithDefaultTrue())
	if err != nil {
		return err
	}
	if result {
		return login(ctx, out)
	}
	return nil
}
