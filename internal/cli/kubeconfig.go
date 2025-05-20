package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/kubeconfig"
	"github.com/nais/cli/internal/naisdevice"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func kubeconfigCommand(rootFlags *root.Flags) *cobra.Command {
	cmdFlags := &kubeconfig.Flags{Flags: rootFlags}
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Create a kubeconfig file for connecting to available clusters.",
		Long: `Create a kubeconfig file for connecting to available clusters

This requires that you have the gcloud command line tool installed, configured and logged in using:
"nais login"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := gcp.ValidateAndGetUserLogin(cmd.Context(), false); err != nil {
				return err
			}

			if mightBeWSL() {
				fmt.Println("Skipping naisdevice check in WSL. Assuming it's connected and ready to go.")
			} else {
				status, err := naisdevice.GetStatus(cmd.Context())
				if err != nil {
					return err
				}

				if !naisdevice.IsConnected(status) {
					return fmt.Errorf("you need to be connected with naisdevice before using this command")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdFlags.Verbose, _ = cmd.Flags().GetBool("verbose")
			return kubeconfig.Run(cmd.Context(), cmdFlags)
		},
	}
	cmd.Flags().StringSliceVarP(&cmdFlags.Exclude, "exclude", "e", nil, "Exclude `CLUSTER` from kubeconfig. Can be repeated.")
	cmd.Flags().BoolVarP(&cmdFlags.Overwrite, "overwrite", "o", false, "Overwrite existing kubeconfig entries if conflicts are found.")
	cmd.Flags().BoolVarP(&cmdFlags.Clear, "clear", "c", false, "Clear existing kubeconfig.")

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
