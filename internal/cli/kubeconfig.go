package cli

import (
	"fmt"
	"os"
	"strings"

	kubeconfigcmd "github.com/nais/cli/internal/kubeconfig"
	"github.com/spf13/cobra"
)

func kubeconfig() *cobra.Command {
	cmdFlags := kubeconfigcmd.Flags{}
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Create a kubeconfig file for connecting to available clusters",
		Long: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			/*
				TODO: add validation

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
			*/
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdFlags.Verbose, _ = cmd.Flags().GetBool("verbose")
			return kubeconfigcmd.Run(cmd.Context(), cmdFlags)
		},
	}
	cmd.Flags().StringSliceVarP(&cmdFlags.Exclude, "exclude", "e", nil, "Exclude clusters from cmd. Can be specified as a comma separated list")
	cmd.Flags().BoolVarP(&cmdFlags.Overwrite, "overwrite", "o", false, "Overwrite existing kubeconfig data if conflicts are found")
	cmd.Flags().BoolVarP(&cmdFlags.Clear, "clear", "c", false, "Clear existing kubeconfig before writing new data")

	return cmd
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
