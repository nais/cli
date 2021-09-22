package cmd

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/consts"
	"github.com/nais/nais-cli/pkg/secret"
	"github.com/spf13/cobra"
	"strings"
)

var GetCmd = &cobra.Command{
	Use:   "get [args] [flags]",
	Short: "Return the preferred config format from a protected secret and generate files to tmp folder",
	Example: `nais aiven get secret-name namespace | nais aiven get secret-name namespace -d ./config | 
nais aiven get secret-name namespace -c kcat | nais aiven get secret-name namespace -c .env | 
 nais aiven get secret-name namespace -c all`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 2 {
			return fmt.Errorf("missing reqired arguments: %s, %s", SecretNameFlag, NamespaceFlag)
		}

		secretName := strings.TrimSpace(args[0])
		namespace := strings.TrimSpace(args[1])

		configType, err := helpers.GetString(cmd, ConfigFlag, false)
		if err != nil {
			return fmt.Errorf("getting %s: %s", ConfigFlag, err)
		}

		if configType != consts.EnvironmentConfigurationType && configType != consts.AllConfigurationType && configType != consts.KCatConfigurationType {
			return fmt.Errorf("valid args: %s | %s | %s", consts.EnvironmentConfigurationType, consts.KCatConfigurationType, consts.AllConfigurationType)
		}

		dest, err := helpers.GetString(cmd, DestFlag, false)
		if err != nil {
			return fmt.Errorf("getting %s: %s", DestFlag, err)
		}
		secret.ExtractAndGenerateConfig(configType, dest, secretName, namespace)
		return nil
	},
}
