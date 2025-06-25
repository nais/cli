package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/kubeconfig/command/flag"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/nais/cli/pkg/cli"
)

func Kubeconfig(rootFlags *root.Flags) *cli.Command {
	flags := &flag.Kubeconfig{Flags: rootFlags}
	return &cli.Command{
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

			if !mightBeWSL() && !naisdevice.IsConnected(ctx) {
				return fmt.Errorf("you need to be connected with naisdevice before using this command")
			}

			return nil
		},
		RunFunc: func(ctx context.Context, out cli.Output, _ []string) error {
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

// mightBeWSL checks if the current environment is likely to be WSL (Windows Subsystem for Linux).
// https://superuser.com/a/1749811
func mightBeWSL() bool {
	if env := os.Getenv("WSL_DISTRO_NAME"); env != "" {
		return true
	}

	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}

	if b, err := os.ReadFile("/proc/version"); err == nil {
		if strings.Contains(string(b), "Microsoft") {
			return true
		}
	}

	return false
}
