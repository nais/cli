package cli

import (
	"github.com/nais/cli/internal/debug"
	"github.com/nais/cli/internal/debug/tidy"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func debugCommand(rootFlags *root.Flags) *cobra.Command {
	cmdFlags := &debug.Flags{Flags: rootFlags}
	cmd := &cobra.Command{
		Use:   "debug APP_NAME",
		Short: "Create and attach to a debug container.",
		Long: `Create and attach to a debug container

When flag "--copy" is set, the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

To debug a live pod, run the command without the "--copy" flag.

You can only reconnect to the debug session if the pod is running.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return debug.Run(args[0], cmdFlags)
		},
	}

	defaultContext, defaultNamespace := k8s.GetDefaultContextAndNamespace()

	cmd.Flags().StringVar(&cmdFlags.Context, "context", defaultContext, "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	cmd.Flags().StringVarP(&cmdFlags.Namespace, "namespace", "n", defaultNamespace, "The kubernetes `NAMESPACE` to use. Defaults to current namespace.")
	cmd.Flags().BoolVar(&cmdFlags.Copy, "copy", false, "Create a copy of the pod with a debug container. The original pod remains running and unaffected.")
	cmd.Flags().BoolVarP(&cmdFlags.ByPod, "by-pod", "b", false, "Attach to a specific `BY-POD` in a workload.")

	tidyCmdFlags := &tidy.Flags{Flags: rootFlags}
	tidyCmd := &cobra.Command{
		Use:   "tidy APP_NAME",
		Short: "Clean up debug containers and debug pods.",
		Long: `Remove debug containers created by the "nais debug" command

Set the "--copy" flag to delete copy pods.`,
		RunE: func(_ *cobra.Command, args []string) error {
			return tidy.Run(args[0], tidyCmdFlags)
		},
	}
	tidyCmd.Flags().StringVar(&tidyCmdFlags.Context, "context", defaultContext, "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	tidyCmd.Flags().StringVarP(&tidyCmdFlags.Namespace, "namespace", "n", defaultNamespace, "The kubernetes `NAMESPACE` to use. Defaults to current namespace.")
	tidyCmd.Flags().BoolVar(&tidyCmdFlags.Copy, "copy", false, "Delete the copy of the original pod. The original pod remains running and unaffected.")

	cmd.AddCommand(tidyCmd)

	return cmd
}
