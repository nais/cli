package cmd

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/aiven"
	aivenclient "github.com/nais/nais-cli/pkg/client"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

const (
	KafkaNavDev             = "nav-dev"
	KafkaNavProd            = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var aivenCommand = &cobra.Command{
	Use:   "aiven [command] [args] [flags]",
	Short: "Create a protected & time-limited aivenApplication",
	Long:  `This command will apply a aivenApplication based on information given and aivenator creates a set of credentials`,
	Example: `nais aiven username namespace | nais aiven username namespace -p nav-dev |
nais aiven username namespace -e 10 | nais aiven username namespace -s some-secret-name`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 2 {
			return fmt.Errorf("%s %s %s : reqired arguments", cmd.CommandPath(), UsernameFlag, NamespaceFlag)
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

		aivenConfig := aiven.SetupAiven(aivenclient.SetupClient(), username, namespace, pool, secretName, expiry)
		aivenApp, err := aivenConfig.GenerateApplication()
		if err != nil {
			return fmt.Errorf("an error occurred generating aivenApplication %s", err)
		}
		log.Default().Printf("use: '%s get %s %s'.", cmd.CommandPath(), aivenApp.Spec.SecretName, aivenApp.Namespace)
		return nil
	},
}
