package command

import (
	"context"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type flags struct {
	*naistrix.GlobalFlags
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
}

func Login(parentFlags *naistrix.GlobalFlags) *naistrix.Command {
	flags := &flags{GlobalFlags: parentFlags}
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
		Group:       auth.GroupName,
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			return gcloud.Login(ctx, out, flags.IsVerbose())
		},
	}
}
