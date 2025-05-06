package cli2

import "github.com/spf13/cobra"

func debugcmd() *cobra.Command {
	debug := &cobra.Command{
		Use:   "debug [app]",
		Short: "Create and attach to a debug container for a given `app`",
		Long: "Create and attach to a debug pod or container. \n" +
			"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
			"allowing you to troubleshoot without affecting the live pod.\n" +
			"To debug a live pod, run the command without the '--copy' flag.\n" +
			"You can only reconnect to the debug session if the pod is running.",
		// Before: debug.Before,
		// Run:    debug.Action,
	}
	debug.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	debug.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	debug.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")
	debug.Flags().Bool("by-pod", false, "Attach to a specific `BY-POD` in a workload")

	tidy := &cobra.Command{
		Use:   "tidy [app]",
		Short: "Clean up debug containers and debug pods",
		Long:  "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		// Before: tidy.Before,
		// Run:    tidy.Action,
	}
	tidy.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	tidy.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	tidy.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")

	debug.AddCommand(tidy)

	return debug
}
