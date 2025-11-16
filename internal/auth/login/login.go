package login

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
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
}

func LoginDeprecated(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &loginFlags{Auth: &flag.Auth{GlobalFlags: parentFlags}}
	return &naistrix.Command{
		Name:  "login",
		Title: "Log in to the Nais platform.",
		Flags: flags,
		Examples: []naistrix.Example{
			{
				Description: "Log in to the Nais platform using gcloud.",
			},
			{
				Description: "Log in to the Nais platform using login.nais.io.",
				Command:     "-n",
			},
		},
		Description: `Uses "gcloud auth login --update-adc" by default.`,
		Deprecated: naistrix.DeprecatedWithReplacementFunc(func(context.Context, *naistrix.Arguments) []string {
			cmd := []string{"auth", "login"}
			if flags.Nais {
				cmd = append(cmd, "-n")
			}

			return cmd
		}),
	}
}

func Login(parentFlags *flag.Auth) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	return &naistrix.Command{
		Name:  "login",
		Title: "Log in to the Nais platform.",
		Examples: []naistrix.Example{
			{
				Description: "Log in to the Nais platform using gcloud.",
			},
			{
				Description: "Log in to the Nais platform using login.nais.io.",
				Command:     "-n",
			},
		},
		Description: `Uses "gcloud auth login --update-adc" by default.`,
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			return gcloud.Login(ctx, out, flags.IsVerbose())
		},
	}
}
