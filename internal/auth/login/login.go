package command

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/auth"
	"github.com/nais/cli/v2/internal/gcloud"
	"github.com/nais/cli/v2/internal/naisapi"
	"github.com/nais/cli/v2/internal/root"
)

type flags struct {
	*root.Flags
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
}

func Login(rootFlags *root.Flags) *cli.Command {
	flags := &flags{Flags: rootFlags}
	return &cli.Command{
		Name:  "login",
		Title: "Log in to the Nais platform.",
		Examples: []cli.Example{
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
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
			if flags.Nais {
				return naisapi.Login(ctx, out)
			}

			return gcloud.Login(ctx, out)
		},
	}
}
