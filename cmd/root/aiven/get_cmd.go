package aiven

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/aiven/secret"
	"github.com/nais/cli/pkg/aiven/services"
	"github.com/spf13/cobra"
	"strings"
)

var getCmd = &cobra.Command{
	Use:   "get [args] [flags]",
	Short: "Generate preferred config format to '/tmp' folder",
	Example: `nais aiven get kafka secret-name namespace | nais aiven get kafka secret-name namespace -c kcat | 
nais aiven get kafka secret-name namespace -c .env | nais aiven get kafka secret-name namespace -c all`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("missing required arguments: %v, %v, %v", cmd.ServiceFlag, cmd.SecretNameFlag, cmd.NamespaceFlag)
		}

		service, err := services.ServiceFromString(strings.TrimSpace(args[0]))
		if err != nil {
			return err
		}
		secretName := strings.TrimSpace(args[1])
		namespace := strings.TrimSpace(args[2])

		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		err = secret.ExtractAndGenerateConfig(service, secretName, namespace)
		if err != nil {
			return fmt.Errorf("retrieve secret and generating config: %w", err)
		}
		return nil
	},
}
