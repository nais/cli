package command

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/urfave/cli/v2"
)

func Login() *cli.Command {
	return &cli.Command{
		Name:        "login",
		Usage:       "Login using Google Auth.",
		Description: "This is a wrapper around gcloud auth login --update-adc.",
		Action: func(context *cli.Context) error {
			return gcp.Login(context.Context)
		},
	}
}
