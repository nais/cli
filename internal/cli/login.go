package cli

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/spf13/cobra"
)

func login() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login using Google Auth.",
		Long:  "This is a wrapper around gcloud auth login --update-adc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gcp.Run(cmd.Context())
		},
	}
}
