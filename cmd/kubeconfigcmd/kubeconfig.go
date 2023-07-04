package kubeconfigcmd

import (
	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/kubeconfig"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "kubeconfig",
		Usage: "Create a kubeconfig file for connecting to available clusters",
		Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "email",
			},
			&cli.BoolFlag{
				Name:    "overwrite",
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:    "clear",
				Aliases: []string{"c"},
			},
			&cli.BoolFlag{
				Name:    "include-onprem",
				Aliases: []string{"io"},
			},
			&cli.BoolFlag{
				Name: "include-ci",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context, false)
		},
		Action: func(context *cli.Context) error {
			email := context.String("email")
			overwrite := context.Bool("overwrite")
			clear := context.Bool("clear")
			includeOnprem := context.Bool("include-onprem")
			includeCi := context.Bool("include-ci")
			verbose := context.Bool("verbose")

			if email == "" {
				var err error
				email, err = gcp.GetActiveUserEmail(context.Context)
				if err != nil {
					return err
				}
			}

			return kubeconfig.CreateKubeconfig(context.Context, email, overwrite, clear, includeOnprem, includeCi, verbose)
		},
	}
}
