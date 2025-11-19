package printaccesstoken

import (
	"context"

	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type printAccessTokenFlags struct {
	*flag.Auth
	Nais bool `name:"nais" short:"n" usage:"Print token from login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais auth login --nais\"."`
}

func PrintAccessToken(parentFlags *flag.Auth) *naistrix.Command {
	flags := &printAccessTokenFlags{Auth: parentFlags}
	return &naistrix.Command{
		Name:        "print-access-token",
		Title:       "Print current access token",
		Description: "Print the last fetched access token",
		Aliases:     []string{"token"},
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if flags.Nais {
				user, err := naisapi.GetAuthenticatedUser(ctx)
				if err != nil {
					return err
				}
				token, err := user.AccessToken()
				if err != nil {
					return err
				}
				out.Println(token)
				return nil
			}

			return gcloud.PrintAccessToken(ctx, out)
		},
	}
}
