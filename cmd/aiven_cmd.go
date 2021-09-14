package cmd

import (
	"fmt"
	"github.com/nais/nais-d/cmd/helpers"
	"github.com/nais/nais-d/pkg/aiven"
	"github.com/spf13/cobra"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var aivenCommand = &cobra.Command{
	Use:   "aiven [command] [args] [flags]",
	Short: "Create a aivenApplication to your cluster",
	Long:  `This command will apply a aivenApplication based on information given and avienator will create a set of credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {


		if len(args) != 2 {
			return fmt.Errorf("%s %s %s : reqired arguments", cmd.CommandPath(), UsernameFlag, TeamFlag)
		}
		username := args[0]
		team := args[1]

		pool, _ := helpers.GetString(cmd, PoolFlag, false)
		if pool != KafkaNavDev && pool != KafkaNavProd && pool != KafkaNavIntegrationTest {
			return fmt.Errorf("valid values for '--%s': %s | %s | %s", PoolFlag, KafkaNavDev, KafkaNavProd, KafkaNavIntegrationTest)
		}

		expiry, err := cmd.Flags().GetInt(ExpireFlag)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		secretName, err := helpers.GetString(cmd, SecretNameFlag, false)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		aivenConfig := aiven.SetupAivenConfiguration(
			aiven.AivenProperties{
				Username:   username,
				Namespace:  team,
				Pool:       pool,
				SecretName: secretName,
				Expiry:     expiry,
			},
		)
		if err := aivenConfig.GenerateApplication(); err != nil {
			return fmt.Errorf("an error occurred generating aivenApplication: %s", err)
		}
		return nil
	},
}
