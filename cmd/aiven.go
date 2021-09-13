package cmd

import (
	"fmt"
	"github.com/nais/debuk/client"
	"github.com/nais/debuk/cmd/helpers"
	"github.com/nais/debuk/config"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var aivenCommand = &cobra.Command{
	Use:   "aiven [ARGS] [FLAGS]",
	Short: "Create a aivenApplication to your cluster",
	Long:  `This command will apply a aivenApplication based on information given and avienator will create a set of credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 2 {
			return fmt.Errorf("username and team is reqired")
		}

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

		dest, err := helpers.GetString(cmd, DestFlag, "", false)
		if err != nil {
			return fmt.Errorf("getting %s: %s", DestFlag, err)
		}

		expiry, err := cmd.Flags().GetInt(ExpireFlag)
		secretName, err := helpers.GetString(cmd, SecretNameFlag, "", false)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		aivenConfig := config.SetupAivenConfiguration(
			client,
			config.AivenProperties{
				Username:   username,
				Namespace:  namespace.Name,
				Pool:       pool,
				Dest:       dest,
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
