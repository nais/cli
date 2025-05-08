package cli

import (
	debugcmd "github.com/nais/cli/internal/debug"
	tidycmd "github.com/nais/cli/internal/debug/tidy"
	"github.com/spf13/cobra"
)

func debug() *cobra.Command {
	cmdFlags := debugcmd.Flags{}
	cmd := &cobra.Command{
		Use:   "debug app",
		Short: "Create and attach to a debug container for a given `app`",
		Long: "Create and attach to a debug pod or container. \n" +
			"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
			"allowing you to troubleshoot without affecting the live pod.\n" +
			"To debug a live pod, run the command without the '--copy' flag.\n" +
			"You can only reconnect to the debug session if the pod is running.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugcmd.Run(args[0], cmdFlags)
		},
	}
	cmd.Flags().StringVarP(&cmdFlags.Context, "context", "c", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	cmd.Flags().StringVarP(&cmdFlags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	cmd.Flags().BoolVarP(&cmdFlags.Copy, "copy", "C", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")
	cmd.Flags().BoolVarP(&cmdFlags.ByPod, "by-pod", "b", false, "Attach to a specific `BY-POD` in a workload")

	tidyCmdFlags := tidycmd.Flags{}
	tidyCmd := &cobra.Command{
		Use:   "tidy app",
		Short: "Clean up debug containers and debug pods",
		Long:  "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tidycmd.Run(args[0], tidyCmdFlags)
		},
	}
	tidyCmd.Flags().StringVarP(&tidyCmdFlags.Context, "context", "c", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	tidyCmd.Flags().StringVarP(&tidyCmdFlags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	tidyCmd.Flags().BoolVarP(&tidyCmdFlags.Copy, "copy", "C", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")

	cmd.AddCommand(tidyCmd)

	return cmd
}
