package aiven

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/aiven/client"
	"github.com/nais/cli/pkg/aiven/services"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var createCmd = &cobra.Command{
	Use:   "create [args] [flags]",
	Short: "Creates an protected and time-limited 'AivenApplication'",
	Long:  `Creates an 'AivenApplication' based on input`,
	Example: `nais aiven create kafka username namespace | nais aiven kafka create username namespace -p kafka-pool |
nais aiven create opensearch username namespace -i soknad -s preferred-secret-name | 
nais aiven opensearch create username namespace -i soknad -a read`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("missing required arguments: %v, %v, %v", cmd.ServiceFlag, cmd.UsernameFlag, cmd.NamespaceFlag)
		}

		service, err := services.ServiceFromString(strings.TrimSpace(args[0]))
		if err != nil {
			return fmt.Errorf("%v\nvalid values for %v: %v", err, cmd.ServiceFlag, services.ValidServices())
		}
		username := strings.TrimSpace(args[1])
		namespace := strings.TrimSpace(args[2])

		poolFlag, err := cmd.GetString(command, cmd.PoolFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %v", err)
		}
		pool, err := services.KafkaPoolFromString(poolFlag)
		if err != nil && service.Is(&services.Kafka{}) {
			return fmt.Errorf("valid values for '-%v': %v",
				cmd.PoolFlag,
				strings.Join(services.KafkaPools, " | "))
		}

		expiry, err := cmd.GetInt(command, cmd.ExpireFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %v", err)
		}

		secretName, err := cmd.GetString(command, cmd.SecretNameFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %v", err)
		}

		instance, err := cmd.GetString(command, cmd.InstanceFlag, service.Is(&services.OpenSearch{}))
		if err != nil {
			return fmt.Errorf("flag: %v", err)
		}

		accessFlag, err := cmd.GetString(command, cmd.AccessFlag, false)
		if err != nil {
			return fmt.Errorf("flag: %v", err)
		}
		access, err := services.OpenSearchAccessFromString(accessFlag)
		if err != nil && service.Is(&services.OpenSearch{}) {
			return fmt.Errorf("valid values for '-%v': %v",
				cmd.AccessFlag,
				strings.Join(services.KafkaPools, " | "))
		}

		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		aivenConfig := aiven.Setup(client.SetupClient(), service, username, namespace, secretName, instance, pool, access, expiry)
		aivenApp, err := aivenConfig.GenerateApplication()
		if err != nil {
			return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
		}
		log.Default().Printf("use: 'nais aiven get %v %v %v' to generate configuration secrets.", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)
		return nil
	},
}
