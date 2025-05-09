package cli

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func login(*root.Flags) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login using Google Auth.",
		Long:  `This is a wrapper around "gcloud auth login --update-adc"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return gcp.Run(cmd.Context())
		},
	}
}
