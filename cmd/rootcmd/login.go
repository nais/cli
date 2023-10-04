package rootcmd

import (
	"github.com/nais/cli/pkg/gcp"
	"github.com/urfave/cli/v2"
)

func loginCommand() *cli.Command {
	return &cli.Command{
		Name:        "login",
		Usage:       "Login using Google Auth.",
		Description: "This is a wrapper around gcloud auth application-default login",
		Action: func(context *cli.Context) error {
			return gcp.Login(context.Context)
		},
	}
}
