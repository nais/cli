package command

import (
	"context"

	"github.com/nais/cli/internal/gcp"
	"github.com/urfave/cli/v3"
)

func Login() *cli.Command {
	return &cli.Command{
		Name:        "login",
		Usage:       "Login using Google Auth.",
		Description: "This is a wrapper around gcloud auth login --update-adc.",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return gcp.Login(ctx)
		},
	}
}
