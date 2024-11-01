package rootcmd

import (
	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

func loginCommand() *cli.Command {
	return &cli.Command{
		Name:        "login",
		Usage:       "Login using Google Auth.",
		Description: "This is a wrapper around gcloud auth login --update-adc.",
		Action: func(context *cli.Context) error {
			metrics.AddOne("login_total")
			return gcp.Login(context.Context)
		},
	}
}
