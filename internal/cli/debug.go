package cli

import "github.com/spf13/cobra"

func debug() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug [app]",
		Short: "Create and attach to a debug container for a given `app`",
		Long: "Create and attach to a debug pod or container. \n" +
			"When flag '--copy' is set, the command can be used to debug a copy of the original pod, \n" +
			"allowing you to troubleshoot without affecting the live pod.\n" +
			"To debug a live pod, run the command without the '--copy' flag.\n" +
			"You can only reconnect to the debug session if the pod is running.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	cmd.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	cmd.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")
	cmd.Flags().Bool("by-pod", false, "Attach to a specific `BY-POD` in a workload")

	tidy := &cobra.Command{
		Use:   "tidy [app]",
		Short: "Clean up debug containers and debug pods",
		Long:  "Remove debug containers created by the 'debug' command. To delete copy pods set the '--copy' flag.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	tidy.Flags().String("context", "", "The kubeconfig `CONTEXT` to use. Defaults to current context.")
	tidy.Flags().String("namespace", "", "The kubernetes `NAMESPACE` to use. Defaults to current namespace in kubeconfig.")
	tidy.Flags().Bool("copy", false, "To create or delete a 'COPY' of pod with a debug container. The original pod remains running and unaffected")

	cmd.AddCommand(tidy)

	return cmd
}
