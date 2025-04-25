package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

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

func Kubeconfig() *cli.Command {
	return &cli.Command{
		Name:  "kubeconfig",
		Usage: "Create a kubeconfig file for connecting to available clusters",
		Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "overwrite",
				Usage:   "Overwrite existing kubeconfig data if conflicts are found",
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:    "clear",
				Usage:   "Clear existing kubeconfig before writing new data",
				Aliases: []string{"c"},
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Usage:   "Exclude clusters from kubeconfig. Can be specified multiple times or as a comma separated list",
				Aliases: []string{"e"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if _, err := gcp.ValidateAndGetUserLogin(ctx, false); err != nil {
				return ctx, err
			}

			if mightBeWSL() {
				fmt.Println("Skipping naisdevice check in WSL. Assuming it's connected and ready to go.")
			} else {
				status, err := naisdevice.GetStatus(ctx)
				if err != nil {
					return ctx, err
				}

				if !naisdevice.IsConnected(status) {
					metrics.AddOne(ctx, "kubeconfig_connect_error_total")
					return ctx, fmt.Errorf("you need to be connected with naisdevice before using this command")
				}
			}

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			clear := cmd.Bool("clear")
			exclude := cmd.StringSlice("exclude")
			overwrite := cmd.Bool("overwrite")
			verbose := cmd.Bool("verbose")

			email, err := gcp.GetActiveUserEmail(ctx)
			if err != nil {
				return err
			}

			return kubeconfig.CreateKubeconfig(ctx, email,
				kubeconfig.WithOverwriteData(overwrite),
				kubeconfig.WithFromScratch(clear),
				kubeconfig.WithExcludeClusters(exclude),
				kubeconfig.WithOnpremClusters(true),
				kubeconfig.WithVerboseLogging(verbose))
		},
	}
}
