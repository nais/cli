package cmd

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/aiven"
	"github.com/nais/nais-cli/pkg/client"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var CreateCmd = &cobra.Command{
	Use:   "create [args] [flags]",
	Short: "Creates a protected and time-limited 'AivenApplication'",
	Long:  `This command will create an 'AivenApplication' based on input`,
	Example: `nais aiven create username namespace | nais aiven create username namespace -p kafka-pool |
nais aiven create username namespace -e 10 | nais aiven create username namespace | 
nais aiven create username namespace -s preferred-secret-name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("missing reqired arguments: %s, %s", UsernameFlag, NamespaceFlag)
		}

		username := strings.TrimSpace(args[0])
		namespace := strings.TrimSpace(args[1])

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

		aivenConfig := aiven.SetupAiven(client.SetupClient(), username, namespace, pool, secretName, expiry)
		aivenApp, err := aivenConfig.GenerateApplication()
		if err != nil {
			return fmt.Errorf("an error occurred generating aivenApplication %s", err)
		}
		log.Default().Printf("use: '%s get %s %s'.", cmd.CommandPath(), aivenApp.Spec.SecretName, aivenApp.Namespace)
		return nil
	},
}
