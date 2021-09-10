package cmd

import (
	"fmt"
	"github.com/nais/debuk/cmd/helpers"
	"github.com/nais/debuk/pkg/generate"
	"github.com/spf13/cobra"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var applyCommand = &cobra.Command{
	Use:   "apply [FLAGS]",
	Short: "Apply a aivenApplication and fetch secrets",
	Long:  `Will apply a aivenApplication based on information given and extract credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {

		username, err := helpers.GetString(cmd, UsernameFlag, true)
		if err != nil {
			return err
		}

		team, err := helpers.GetString(cmd, TeamFlag, true)
		if err != nil {
			return err
		}

		pool, _ := helpers.GetString(cmd, PoolFlag, false)
		if pool != KafkaNavDev && pool != KafkaNavProd && pool != KafkaNavIntegrationTest {
			return fmt.Errorf("valid values for '--%s': %s | %s | %s", PoolFlag, KafkaNavDev, KafkaNavProd, KafkaNavIntegrationTest)
		}

		dest, err := helpers.GetString(cmd, DestFlag, false)
		if err != nil {
			return fmt.Errorf("getting %s: %s", DestFlag, err)
		}

		expiry, err := cmd.Flags().GetInt(ExpireFlag)
		secretName, err := helpers.GetString(cmd, SecretNameFlag, false)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		if err := generate.AivenApplication(username, team, pool, dest, expiry, secretName); err != nil {
			return fmt.Errorf("an error occurred generating aivenApplication: %s", err)
		}
		return nil
	},
}
