package cli

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/nais"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func login(_ *root.Flags) *cobra.Command {
	cmdFlagNais := false

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login using Google Auth.",
		Long:  `This is a wrapper around "gcloud auth login --update-adc"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmdFlagNais {
				return nais.Login(cmd.Context())
			}

			return gcp.Run(cmd.Context())
		},
	}

	cmd.Flags().BoolVarP(&cmdFlagNais, "nais", "n", cmdFlagNais, "Very good description.")
	return cmd
}
