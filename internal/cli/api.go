package cli

import (
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	naisapiproxy "github.com/nais/cli/internal/naisapi/proxy"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func api(rootFlags *root.Flags) *cobra.Command {
	cmdFlags := &naisapi.Flags{
		Flags: rootFlags,
	}
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Interact with Nais API.",
	}

	proxyCmdFlags := &naisapiproxy.Flags{
		Flags:      cmdFlags,
		ListenAddr: "localhost:4242",
	}
	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "Authenticated proxy to do GraphQL requests against the Nais API.",
		Long:  `Starts a proxy server that authenticates requests to the Nais API using your account token.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return naisapiproxy.Run(cmd.Context(), proxyCmdFlags)
		},
	}
	proxyCmd.Flags().StringVarP(&proxyCmdFlags.ListenAddr, "listen", "l", proxyCmdFlags.ListenAddr, "Address the proxy will listen on.")

	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Outputs the Nais API GraphQL schema to stdout.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			schema, err := naisapi.PullSchema(cmd.Context())
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), schema)
			return nil
		},
	}

	teamsCmd := &cobra.Command{
		Use:   "teams",
		Short: "Get a list of your teams.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			teams, err := naisapi.GetUserTeams(cmd.Context())
			if err != nil {
				return err
			}

			if len(teams.Me.(*gql.UserTeamsMeUser).Teams.Nodes) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No teams found.")
				return nil
			}

			for _, team := range teams.Me.(*gql.UserTeamsMeUser).Teams.Nodes {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), team.Team.Slug, "-", team.Team.Purpose)
			}

			return nil
		},
	}

	cmd.AddCommand(
		proxyCmd,
		schemaCmd,
		teamsCmd,
	)
	return cmd
}
