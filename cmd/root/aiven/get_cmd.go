package aiven

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/consts"
	"github.com/nais/cli/pkg/secret"
	"github.com/spf13/cobra"
	"strings"
)

var getCmd = &cobra.Command{
	Use:   "get [args] [flags]",
	Short: "Generate preferred config format to '/tmp' folder",
	Example: `nais aiven get secret-name namespace | nais aiven get secret-name namespace -c kcat | 
nais aiven get secret-name namespace -c .env | nais aiven get secret-name namespace -c all`,
	RunE: func(command *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("missing required arguments: %s, %s", cmd.SecretNameFlag, cmd.NamespaceFlag)
		}

		secretName := strings.TrimSpace(args[0])
		namespace := strings.TrimSpace(args[1])

		configType, err := cmd.GetString(command, cmd.ConfigFlag, false)
		if err != nil {
			return fmt.Errorf("'--%s': %w", cmd.ConfigFlag, err)
		}

		if configType != consts.EnvironmentConfigurationType && configType != consts.AllConfigurationType && configType != consts.KCatConfigurationType {
			return fmt.Errorf("valid values for '--%s': %s, %s, %s",
				cmd.ConfigFlag,
				consts.EnvironmentConfigurationType,
				consts.KCatConfigurationType,
				consts.AllConfigurationType,
			)
		}

		// workaround https://github.com/spf13/cobra/issues/340
		command.SilenceUsage = true

		err = secret.ExtractAndGenerateConfig(configType, secretName, namespace)
		if err != nil {
			return fmt.Errorf("retrieve secret and generating config: %w", err)
		}
		return nil
	},
}
