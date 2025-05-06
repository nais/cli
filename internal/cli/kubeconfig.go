package cli

import (
	"github.com/spf13/cobra"
)

func kubeconfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Create a kubeconfig file for connecting to available clusters",
		Long: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().StringSlice("exclude", nil, "Exclude clusters from cmd. Can be specified as a comma separated list")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing kubeconfig data if conflicts are found")
	cmd.Flags().Bool("clear", false, "Clear existing kubeconfig before writing new data")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	return cmd
}
