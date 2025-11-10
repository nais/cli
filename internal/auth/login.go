package auth

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
	flags := &flag.Auth{GlobalFlags: parentFlags}
	return Login(flags, Deprecated)
}

func Login(parentFlags *flag.Auth, modifiers ...func(*naistrix.Command)) *naistrix.Command {
	flags := &loginFlags{Auth: parentFlags}
	cmd := &naistrix.Command{
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
	for _, modifier := range modifiers {
		modifier(cmd)
	}
	return cmd
}
