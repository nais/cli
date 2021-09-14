package cmd

import (
	"fmt"
	"github.com/nais/nais-d/client"
	"github.com/nais/nais-d/cmd/helpers"
	"github.com/nais/nais-d/pkg/aiven"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var aivenCommand = &cobra.Command{
	Use:   "aiven [commands] [args] [flags]",
	Short: "Create a aivenApplication to your cluster",
	Long:  `This command will apply a aivenApplication based on information given and avienator will create a set of credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {

		client := client.StandardClient()

		username, err := helpers.GetString(cmd, UsernameFlag, args[0], true)
		if err != nil {
			return err
		}

		team, err := helpers.GetString(cmd, TeamFlag, args[1], true)
		if err != nil {
			return err
		}

		namespace, err := client.CoreV1().Namespaces().Get(team, metav1.GetOptions{})
		if err != nil {
			return err
		}

		pool, _ := helpers.GetString(cmd, PoolFlag, "", false)
		if pool != KafkaNavDev && pool != KafkaNavProd && pool != KafkaNavIntegrationTest {
			return fmt.Errorf("valid values for '--%s': %s | %s | %s", PoolFlag, KafkaNavDev, KafkaNavProd, KafkaNavIntegrationTest)
		}

		expiry, err := cmd.Flags().GetInt(ExpireFlag)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		secretName, err := helpers.GetString(cmd, SecretNameFlag, "", false)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		aivenConfig := aiven.SetupAivenConfiguration(
			client,
			aiven.AivenProperties{
				Username:   username,
				Namespace:  namespace.Name,
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
