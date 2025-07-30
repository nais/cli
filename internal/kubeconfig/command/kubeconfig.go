package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/kubeconfig/command/flag"
	"github.com/nais/cli/internal/root"
	"github.com/nais/naistrix"
)

func Kubeconfig(rootFlags *root.Flags) *naistrix.Command {
	flags := &flag.Kubeconfig{
		Flags:     rootFlags,
		Overwrite: true,
	}
	return &naistrix.Command{
		Name:  "kubeconfig",
		Title: "Create a kubeconfig file for connecting to available clusters.",
		Description: heredoc.Doc(`
			This requires that you have the gcloud command line tool installed, configured and logged in using:

			"nais login"
		`),
		Flags: flags,
		ValidateFunc: func(ctx context.Context, args []string) error {
			if _, err := gcloud.ValidateAndGetUserLogin(ctx, false); err != nil {
				return err
			}

			return nil
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
			email, err := gcloud.GetActiveUserEmail(ctx)
			if err != nil {
				return err
			}

			return kubeconfig.CreateKubeconfig(
				ctx,
				email,
				out,
				kubeconfig.WithOverwriteData(flags.Overwrite),
				kubeconfig.WithFromScratch(flags.Clear),
				kubeconfig.WithExcludeClusters(flags.Exclude),
				kubeconfig.WithOnpremClusters(true),
				kubeconfig.WithVerboseLogging(flags.IsVerbose()),
			)
		},
	}
}
