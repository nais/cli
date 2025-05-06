package cli2

import (
	"fmt"

	"github.com/nais/cli/internal/aiven/create"
	"github.com/spf13/cobra"
)

func aivencmd() *cobra.Command {
	aivenCmd := &cobra.Command{
		Use:   "aiven",
		Short: "Command used for management of AivenApplication",
	}

	createFlags := create.Flags{}
	createArgs := func(args []string) create.Args {
		return create.Args{
			Service:   args[0],
			Username:  args[1],
			Namespace: args[2],
		}
	}

	createCmd := &cobra.Command{
		Use:   "create service username namespace",
		Short: "Creates a protected and time-limited AivenApplication",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return create.Action(cmd.Context(), createArgs(args), createFlags)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if createFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			if createFlags.Pool == "" {
				return fmt.Errorf("--pool must not be empty")
			}

			return create.Validate(cmd.Context(), createArgs(args), createFlags)
		},
	}

	createCmd.Flags().UintVarP(&createFlags.Expire, "expire", "e", 1, "Days until credential expires")
	createCmd.Flags().StringVarP(&createFlags.Pool, "pool", "p", "nav-dev", "Kafka pool")
	createCmd.Flags().StringVarP(&createFlags.Secret, "secret", "s", "", "Secret name to store credentials. Will be generated if not provided")
	createCmd.Flags().StringVarP(&createFlags.Instance, "instance", "i", "", "Instance name")
	createCmd.Flags().StringVarP(&createFlags.Access, "access", "a", "", "Access name")

	aivenCmd.AddCommand(createCmd)

	getCmd := &cobra.Command{
		Use:   "get service username namespace",
		Short: "Generate preferred config format to '/tmp' folder",
		// Before:    aivenget.Before,
		// Run:       aivenget.Action,
	}
	aivenCmd.AddCommand(getCmd)

	tidyCmd := &cobra.Command{
		Use:   "tidy",
		Short: "Clean up /tmp/aiven-secret-* made by nais-cli",
		Long: `Remove '/tmp' folder '$TMPDIR' and files created by the aiven command
	Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'`,
		// Run: aiventidy.Action,
	}
	aivenCmd.AddCommand(tidyCmd)

	return aivenCmd
}
