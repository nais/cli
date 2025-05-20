package cli

import (
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	aivencreate "github.com/nais/cli/internal/aiven/create"
	aivencreatekafka "github.com/nais/cli/internal/aiven/create/kafka"
	aivencreateopensearch "github.com/nais/cli/internal/aiven/create/opensearch"
	"github.com/nais/cli/internal/aiven/get"
	"github.com/nais/cli/internal/aiven/tidy"
	"github.com/nais/cli/internal/root"
	"github.com/spf13/cobra"
)

func aivenCommand(rootFlags *root.Flags) *cobra.Command {
	cmdFlags := &aiven.Flags{Flags: rootFlags}
	cmd := &cobra.Command{
		Use:   "aiven",
		Short: "Manage Aiven services.",
	}

	createCmdFlags := &aivencreate.Flags{Flags: cmdFlags}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Grant a user access to an Aiven service.",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if createCmdFlags.Expire > 30 {
				return fmt.Errorf("--expire must be less than %v days", 30)
			}

			return nil
		},
	}
	createCmd.PersistentFlags().UintVarP(&createCmdFlags.Expire, "expire", "e", 1, "Number of `DAYS` until the generated credentials expire.")
	createCmd.PersistentFlags().StringVarP(&createCmdFlags.Secret, "secret", "s", "", "`NAME` of the Kubernetes secret to store the credentials in. Will be generated if not provided.")

	createArgs := func(args []string) aivencreate.Arguments {
		return aivencreate.Arguments{
			Username:  args[0],
			Namespace: args[1],
		}
	}

	var createKafkaPool string
	createKafkaCmd := &cobra.Command{
		Use:   "kafka USERNAME NAMESPACE",
		Short: "Grant a user access to a Kafka topic.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pool, err := aiven_services.KafkaPoolFromString(createKafkaPool)
			if err != nil {
				return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
			}

			return aivencreatekafka.Run(
				cmd.Context(),
				createArgs(args),
				&aivencreatekafka.Flags{
					Flags: createCmdFlags,
					Pool:  pool,
				},
			)
		},
	}
	createKafkaCmd.Flags().StringVarP(&createKafkaPool, "pool", "p", "nav-dev", "The `NAME` of the pool to create the Kafka instance in.")
	_ = createKafkaCmd.MarkFlagRequired("pool")

	var createOpenSearchAccess, createOpenSearchInstance string
	createOpenSearchCmd := &cobra.Command{
		Use:   "opensearch USERNAME NAMESPACE",
		Short: "Grant a user access to an OpenSearch instance.",
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
				&aivencreateopensearch.Flags{
					Flags:  createCmdFlags,
					Access: access,
				},
			)
		},
	}
	createOpenSearchCmd.Flags().StringVarP(&createOpenSearchAccess, "access", "a", "read", fmt.Sprintf("The access `LEVEL`. Available levels: %s", strings.Join(aiven_services.OpenSearchAccesses, ", ")))
	createOpenSearchCmd.Flags().StringVarP(&createOpenSearchInstance, "instance", "i", "", "The name of the OpenSearch `INSTANCE`.")
	_ = createOpenSearchCmd.MarkFlagRequired("instance")
	_ = createOpenSearchCmd.RegisterFlagCompletionFunc("access", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return aiven_services.OpenSearchAccesses, cobra.ShellCompDirectiveNoFileComp
	})

	createCmd.AddCommand(
		createKafkaCmd,
		createOpenSearchCmd,
	)

	cmd.AddCommand(
		createCmd,
		&cobra.Command{
			Use:   "get SERVICE USERNAME NAMESPACE",
			Short: "Generate preferred config format to '/tmp' folder.",
			Args:  cobra.ExactArgs(3),
			ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
				comps := make([]cobra.Completion, 0)
				if len(args) == 0 {
					comps = append(comps, "kafka", "opensearch")
					comps = cobra.AppendActiveHelp(comps, "Choose the service you want to get.")
				}
				return comps, cobra.ShellCompDirectiveNoFileComp
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
			Short: "Clean up /tmp/aiven-secret-* files made by the Nais CLI.",
			Long: `Clean up /tmp/aiven-secret-* files made by the Nais CLI

Caution - This command will delete all files in "/tmp" folder starting with "aiven-secret-".`,
			RunE: func(*cobra.Command, []string) error {
				return tidy.Run()
			},
		},
	)

	return cmd
}
