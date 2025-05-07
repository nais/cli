package cli

import (
	debugcmd "github.com/nais/cli/internal/debug"
	tidycmd "github.com/nais/cli/internal/debug/tidy"
	"github.com/spf13/cobra"
)

func debug() *cobra.Command {
	flags := debugcmd.Flags{}
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
			return debugcmd.Run(args[0], flags)
		},
	}
	fs := cmd.Flags()
	fs.StringVarP(&flags.Context, "context", "c", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	fs.StringVarP(&flags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	fs.BoolVarP(&flags.Copy, "copy", "C", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")
	fs.BoolVarP(&flags.ByPod, "by-pod", "b", false, "Attach to a specific `BY-POD` in a workload")

	tidyFlags := tidycmd.Flags{}
	tidy := &cobra.Command{
		Use:   "tidy app",
		Short: "Clean up debug containers and debug pods",
		Long:  "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tidycmd.Run(args[0], tidyFlags)
		},
	}

	fs = tidy.Flags()
	fs.StringVarP(&tidyFlags.Context, "context", "c", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	fs.StringVarP(&tidyFlags.Namespace, "namespace", "n", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	fs.BoolVarP(&tidyFlags.Copy, "copy", "C", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")

	cmd.AddCommand(tidy)

	return cmd
}
