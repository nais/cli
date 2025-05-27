package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/kubeconfig/command/flag"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Kubeconfig(rootFlags *root.Flags) *cli.Command {
	flags := &flag.Kubeconfig{Flags: rootFlags}
	return cli.NewCommand("kubeconfig", "Create a kubeconfig file for connecting to available clusters.",
		cli.WithLong(`Create a kubeconfig file for connecting to available clusters

This requires that you have the gcloud command line tool installed, configured and logged in using:
"nais login"`),
		cli.WithFlag("exclude", "e", "Exclude `CLUSTER` from kubeconfig. Can be repeated.", &flags.Exclude),
		cli.WithFlag("overwrite", "o", "Overwrite existing kubeconfig entries if conflicts are found.", &flags.Overwrite),
		cli.WithFlag("clear", "c", "Clear existing kubeconfig.", &flags.Clear),

		cli.WithValidate(func(ctx context.Context, args []string) error {
			if _, err := gcp.ValidateAndGetUserLogin(ctx, false); err != nil {
				return err
			}

			if mightBeWSL() {
				fmt.Println("Skipping naisdevice check in WSL. Assuming it's connected and ready to go.")
			} else {
				if !naisdevice.IsConnected(ctx) {
					return fmt.Errorf("you need to be connected with naisdevice before using this command")
				}
			}

			return nil
		}),
		cli.WithRun(func(ctx context.Context, out output.Output, _ []string) error {
			email, err := gcp.GetActiveUserEmail(ctx)
			if err != nil {
				return err
			}

			return kubeconfig.CreateKubeconfig(
				ctx,
				email,
				kubeconfig.WithOverwriteData(flags.Overwrite),
				kubeconfig.WithFromScratch(flags.Clear),
				kubeconfig.WithExcludeClusters(flags.Exclude),
				kubeconfig.WithOnpremClusters(true),
				kubeconfig.WithVerboseLogging(flags.IsVerbose()),
			)
		}),
	)
}

func mightBeWSL() bool {
	// https://superuser.com/a/1749811
	env := os.Getenv("WSL_DISTRO_NAME")
	if env != "" {
		fmt.Printf("WSL detected: WSL_DISTRO_NAME=%v\n", env)
		return true
	}

	wslInteropPath := "/proc/sys/fs/binfmt_misc/WSLInterop"
	if _, err := os.Stat(wslInteropPath); err == nil {
		fmt.Printf("WSL detected: %q exists\n", wslInteropPath)
		return true
	}

	procVersionPath := "/proc/version"
	if b, err := os.ReadFile(procVersionPath); err == nil {
		if strings.Contains(string(b), "Microsoft") {
			fmt.Printf("WSL detected: %q contains 'Microsoft'\n", procVersionPath)
			return true
		}
	}

	return false
}
