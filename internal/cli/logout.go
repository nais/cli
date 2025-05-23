package cli

import (
	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func logoutCommand(w output.Output, rootFlags *root.Flags) *cobra.Command {
	cmdFlagNais := false
	cmd := &cobra.Command{
		GroupID: authGroup.ID,
		Use:     "logout",
		Short:   "Log out and remove credentials.",
		Long:    "This logs you out of Nais and removes credentials from your local machine.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmdFlagNais {
				return naisapi.Logout(cmd.Context(), w)
			}

			return gcp.Logout(cmd.Context(), w)
		},
	}

	cmd.Flags().BoolVarP(&cmdFlagNais, "nais", "n", cmdFlagNais, "Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\".")
	return cmd
}
