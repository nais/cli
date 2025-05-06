package cli

import (
	"github.com/spf13/cobra"
)

func kubeconfigcmd() *cobra.Command {
	kubeconfig := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Create a kubeconfig file for connecting to available clusters",
		Long: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
			// return kubeconfig.Before( ... )
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
			// return kubeconfig.Action( ... )
		},
	}
	kubeconfig.Flags().StringSlice("exclude", nil, "Exclude clusters from kubeconfig. Can be specified as a comma separated list")
	kubeconfig.Flags().Bool("overwrite", false, "Overwrite existing kubeconfig data if conflicts are found")
	kubeconfig.Flags().Bool("clear", false, "Clear existing kubeconfig before writing new data")
	kubeconfig.Flags().Bool("verbose", false, "Verbose output")
	return kubeconfig
}
