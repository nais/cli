package cli

import (
	"github.com/nais/cli/internal/nais"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func apiProxy(*root.Flags) *cobra.Command {
	listenAddr := "localhost:4242"

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Authenticated proxy to do GraphQL requests to Nais API.",
		Long:  `Starts a proxy server that authenticates requests to the Nais API using your account token.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return nais.RunAPIProxy(cmd.Context(), listenAddr)
		},
	}

	cmd.Flags().StringVarP(&listenAddr, "listen", "l", listenAddr, "Very good description.")
	return cmd
}
