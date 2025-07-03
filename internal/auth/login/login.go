package command

import (
	"context"

	"github.com/nais/cli/v2/internal/auth"
	"github.com/nais/cli/v2/internal/gcloud"
	"github.com/nais/cli/v2/internal/naisapi"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

type flags struct {
	*root.Flags
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
}

func Login(rootFlags *root.Flags) *naistrix.Command {
	flags := &flags{Flags: rootFlags}
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
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			return gcloud.Login(ctx, out)
		},
	}
}
