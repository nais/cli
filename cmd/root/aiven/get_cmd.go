package aiven

import (
	"fmt"
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/consts"
	"github.com/nais/cli/pkg/secret"
	"github.com/spf13/cobra"
	"strings"
)

var ConfigTypes []string

func init() {
	ConfigTypes = []string{
		consts.JavaConfigurationType,
		consts.KCatConfigurationType,
		consts.EnvironmentConfigurationType,
		consts.AllConfigurationType,
	}
}

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

		validConfigType := false
		for _, candidate := range ConfigTypes {
			if candidate == configType {
				validConfigType = true
				break
			}
		}

		if !validConfigType {
			return fmt.Errorf("valid values for '--%s': %s",
				cmd.ConfigFlag,
				strings.Join(ConfigTypes, ", "),
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
