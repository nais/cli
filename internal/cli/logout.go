package cli

import (
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func logoutCommand(rootFlags *root.Flags) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: authGroup.ID,
		Use:     "logout",
		Short:   "Log out and remove credentials.",
		Long:    "This logs you out of Nais and removes credentials from your local machine.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return naisapi.Logout(cmd.Context(), rootFlags)
		},
	}

	return cmd
}
