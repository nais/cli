package cli

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func loginCommand(w output.Output, rootFlags *root.Flags) *cobra.Command {
	cmdFlagNais := false

	cmd := &cobra.Command{
		GroupID: authGroup.ID,
		Use:     "login",
		Short:   "Login to the Nais platform.",
		Long:    `Login to the Nais platform, uses "gcloud auth login --update-adc" by default.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmdFlagNais {
				return naisapi.Login(cmd.Context(), w)
			}

			return gcp.Login(cmd.Context(), w)
		},
	}

	cmd.Flags().BoolVarP(&cmdFlagNais, "nais", "n", cmdFlagNais, "Login using login.nais.io instead of gcloud.")
	return cmd
}
