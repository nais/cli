package aiven

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/client"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var createCmd = &cobra.Command{
	Use:   "create [args] [flags]",
	Short: "Creates an protected and time-limited 'AivenApplication'",
	Long:  `Creates an 'AivenApplication' based on input`,
	Example: `nais aiven create username namespace | nais aiven create username namespace -p kafka-pool |
nais aiven create username namespace -e 10 | nais aiven create username namespace -s preferred-secret-name`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("missing required arguments: %s, %s", cmd.UsernameFlag, cmd.NamespaceFlag)
		}

		username := strings.TrimSpace(args[0])
		namespace := strings.TrimSpace(args[1])

		pool, err := cmd.GetString(command, cmd.PoolFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %s", err)
		}

		if pool != KafkaNavDev && pool != KafkaNavProd && pool != KafkaNavIntegrationTest {
			return fmt.Errorf("valid values for '-%s': %s | %s | %s",
				cmd.PoolFlag,
				KafkaNavDev,
				KafkaNavProd,
				KafkaNavIntegrationTest,
			)
		}

		expiry, err := cmd.GetInt(command, cmd.ExpireFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %s", err)
		}

		secretName, err := cmd.GetString(command, cmd.SecretNameFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %s", err)
		}

		aivenConfig := aiven.SetupAiven(client.SetupClient(), username, namespace, pool, secretName, expiry)
		aivenApp, err := aivenConfig.GenerateApplication()
		if err != nil {
			return fmt.Errorf("an error occurred generating 'AivenApplication': %s", err)
		}
		log.Default().Printf("use: '%s get %s %s' to generate configuration secrets.", "nais aiven", aivenApp.Spec.SecretName, aivenApp.Namespace)
		return nil
	},
}
