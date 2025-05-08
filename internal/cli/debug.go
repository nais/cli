package cli

import (
	debugcmd "github.com/nais/cli/internal/debug"
	tidycmd "github.com/nais/cli/internal/debug/tidy"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func debug(rootFlags root.Flags) *cobra.Command {
	cmdFlags := debugcmd.Flags{Flags: rootFlags}
	cmd := &cobra.Command{
		Use:   "debug APP",
		Short: "Create and attach to a debug container.",
		Long: `Create and attach to a debug container

When flag "--copy" is set, the command can be used to debug a copy of the original pod, allowing you to troubleshoot without affecting the live pod.

To debug a live pod, run the command without the "--copy" flag.

You can only reconnect to the debug session if the pod is running.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return debugcmd.Run(args[0], cmdFlags)
		},
	}
	// TODO: Fetch current context from kubeconfig and set as actual default and mark flag as required
	cmd.Flags().StringVar(&cmdFlags.Context, "context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	// TODO: Fetch current namespace from kubeconfig and set as actual default and mark flag as required
	cmd.Flags().StringVarP(&cmdFlags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace.")
	cmd.Flags().BoolVar(&cmdFlags.Copy, "copy", false, "Create a copy of the pod with a debug container. The original pod remains running and unaffected.")
	cmd.Flags().BoolVarP(&cmdFlags.ByPod, "by-pod", "b", false, "Attach to a specific `BY-POD` in a workload.")

	tidyCmdFlags := tidycmd.Flags{Flags: rootFlags}
	tidyCmd := &cobra.Command{
		Use:   "tidy app",
		Short: "Clean up debug containers and debug pods.",
		Long: `Remove debug containers created by the "debug" command.

To delete copy pods set the "--copy" flag.`,
		RunE: func(_ *cobra.Command, args []string) error {
			return tidycmd.Run(args[0], tidyCmdFlags)
		},
	}
	// TODO: Fetch current context from kubeconfig and set as actual default and mark flag as required
	tidyCmd.Flags().StringVar(&tidyCmdFlags.Context, "context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	// TODO: Fetch current namespace from kubeconfig and set as actual default and mark flag as required
	tidyCmd.Flags().StringVarP(&tidyCmdFlags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace.")
	tidyCmd.Flags().BoolVar(&tidyCmdFlags.Copy, "copy", false, "Delete the copy of the original pod. The original pod remains running and unaffected.")

	cmd.AddCommand(tidyCmd)

	return cmd
}
