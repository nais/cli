package login

import (
	"context"
	"os"

	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type loginFlags struct {
	*flag.Auth
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
	Yes  bool `name:"yes" short:"y" usage:"Automatically answer yes to all prompts."`
}

func Login(parentFlags *flag.Auth) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	return &naistrix.Command{
		Name:            "login",
		Title:           "Log in to the Nais platform.",
		TopLevelAliases: []string{"login"},
		Examples: []naistrix.Example{
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

			if term.IsTerminal(int(os.Stdin.Fd())) { // #nosec G115 -- fd fits in int on all supported platforms
				pterm.Println()
				pterm.Println("Many Nais commands require you to be logged in to both Google and Nais.")
				if flags.Yes {
					return naisapi.Login(ctx, out)
				}
				result, _ := pterm.DefaultInteractiveConfirm.
					WithDefaultValue(true).
					Show("Would you like to also log in to Nais?")
				if result {
					return naisapi.Login(ctx, out)
				}
			}

			return nil
		},
	}
}
