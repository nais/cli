package cmd

import (
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/config"
	"github.com/nais/nais-cli/pkg/secret"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var getCmd = &cobra.Command{
	Use:   "get [args] [flags]",
	Short: "Return the preferred config format from a protected secret and generate files to location",
	Example: `nais aiven get secret-name namespace | nais aiven get secret-name namespace -d ./config | 
nais aiven get secret-name namespace -c kcat | nais aiven get secret-name namespace -c .env | 
 nais aiven get secret-name namespace -c all`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 2 {
			log.Fatalf("%s and %s are reqired arguments", SecretNameFlag, TeamFlag)
		}

		secretName := strings.TrimSpace(args[0])
		team := strings.TrimSpace(args[1])

		configType, err := helpers.GetString(cmd, ConfigFlag, false)
		if err != nil {
			log.Fatalf("getting %s: %s", ConfigFlag, err)
		}

		if configType != config.ENV && configType != config.ALL && configType != config.KCAT {
			log.Fatalf("valid args: %s | %s | %s", config.ENV, config.KCAT, config.ALL)
		}

		dest, err := helpers.GetString(cmd, DestFlag, false)
		if err != nil {
			log.Fatalf("getting %s: %s", DestFlag, err)
		}
		secret.ExtractAndGenerateConfig(configType, dest, secretName, team)
	},
}
