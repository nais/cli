package cli

import (
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven/aiven_services"
	aivencreate "github.com/nais/cli/internal/aiven/create"
	aivencreatekafka "github.com/nais/cli/internal/aiven/create/kafka"
	aivencreateopensearch "github.com/nais/cli/internal/aiven/create/opensearch"
	"github.com/nais/cli/internal/aiven/get"
	"github.com/nais/cli/internal/aiven/tidy"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func aiven(root.Flags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aiven",
		Short: "Command used for management of AivenApplication",
	}

	createCmdFlags := aivencreate.Flags{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a protected and time-limited AivenApplication",
		Args:  cobra.ExactArgs(3),
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if createCmdFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			return nil
		},
	}
	createCmd.PersistentFlags().UintVarP(&createCmdFlags.Expire, "expire", "e", 1, "Days until credential expires")
	createCmd.PersistentFlags().StringVarP(&createCmdFlags.Secret, "secret", "s", "", "Secret name to store credentials. Will be generated if not provided")

	createArgs := func(args []string) aivencreate.Arguments {
		return aivencreate.Arguments{
			Username:  args[0],
			Namespace: args[1],
		}
	}

	var createKafkaPool string
	createKafkaCmd := &cobra.Command{
		Use:   "kafka username namespace",
		Short: "Creates a protected and time-limited AivenApplication",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pool, err := aiven_services.KafkaPoolFromString(createKafkaPool)
			if err != nil {
				return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
			}

			return aivencreatekafka.Run(
				cmd.Context(),
				createArgs(args),
				aivencreatekafka.Flags{
					Flags: createCmdFlags,
					Pool:  pool,
				},
			)
		},
	}
	createKafkaCmd.Flags().StringVarP(&createKafkaPool, "pool", "p", "nav-dev", "Kafka pool")
	_ = createKafkaCmd.MarkFlagRequired("pool")

	var createOpenSearchAccess, createOpenSearchInstance string
	createOpenSearchCmd := &cobra.Command{
		Use:   "opensearch username namespace",
		Short: "Creates a protected and time-limited AivenApplication",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			access, err := aiven_services.OpenSearchAccessFromString(createOpenSearchAccess)
			if err != nil {
				return fmt.Errorf(
					"valid values for access: %v",
					strings.Join(aiven_services.OpenSearchAccesses, ", "),
				)
			}

			return aivencreateopensearch.Run(
				cmd.Context(),
				createArgs(args),
				aivencreateopensearch.Flags{
					Flags:  createCmdFlags,
					Access: access,
				},
			)
		},
	}
	createOpenSearchCmd.Flags().StringVarP(&createOpenSearchAccess, "access", "a", "", "Access name")
	createOpenSearchCmd.Flags().StringVarP(&createOpenSearchInstance, "instance", "i", "", "Instance name")
	_ = createOpenSearchCmd.MarkFlagRequired("access")

	createCmd.AddCommand(
		createKafkaCmd,
		createOpenSearchCmd,
	)

	cmd.AddCommand(
		createCmd,
		&cobra.Command{
			Use:   "get service username namespace",
			Short: "Generate preferred config format to '/tmp' folder",
			Args:  cobra.ExactArgs(3),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// TODO: audocomplete service name (kafka / opensearch)
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				service, err := aiven_services.FromString(args[0])
				if err != nil {
					return err
				}

				return get.Run(cmd.Context(), service, get.Arguments{
					SecretName: args[1],
					Namespace:  args[2],
				})
			},
		},
		&cobra.Command{
			Use:   "tidy",
			Short: "Clean up /tmp/aiven-secret-* made by nais-cli",
			Long: `Remove '/tmp' folder '$TMPDIR' and files created by the aiven command
	Caution - This will delete all files in '/tmp' folder starting with 'aiven-secret-'`,
			RunE: func(*cobra.Command, []string) error {
				return tidy.Run()
			},
		},
	)

	return cmd
}
