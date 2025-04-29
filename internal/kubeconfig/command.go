package kubeconfig

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/metrics"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
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
}

func Action(ctx context.Context, cmd *cli.Command) error {
	email, err := gcp.GetActiveUserEmail(ctx)
	if err != nil {
		return err
	}

	return CreateKubeconfig(
		ctx,
		email,
		WithOverwriteData(cmd.Bool("overwrite")),
		WithFromScratch(cmd.Bool("clear")),
		WithExcludeClusters(cmd.StringSlice("exclude")),
		WithOnpremClusters(true),
		WithVerboseLogging(cmd.Bool("verbose")),
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
