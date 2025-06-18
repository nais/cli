package command

import (
	"context"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/root"
)

type flags struct {
	*root.Flags
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
}

func Login(rootFlags *root.Flags) *cli.Command {
	flags := &flags{Flags: rootFlags}
	return &cli.Command{
		Name:        "login",
		Title:       "Log in to the Nais platform.",
		Description: `Uses "gcloud auth login --update-adc" by default.`,
		Group:       auth.GroupName,
		Flags:       flags,
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			return gcp.Login(ctx, out)
		},
	}
}
