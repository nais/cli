package rootcmd

import (
	"github.com/nais/cli/pkg/nais"
	"github.com/urfave/cli/v2"
)

func loginCommand() *cli.Command {
	return &cli.Command{
		Name:        "login",
		Usage:       "Login using Google Auth.",
		Description: "This is a wrapper around gcloud auth login --update-adc.",
		Action: func(context *cli.Context) error {
			return nais.Login(context.Context)
		},
	}
}
