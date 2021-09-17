package cmd

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/helpers"
	"github.com/nais/nais-cli/pkg/config"
	"github.com/nais/nais-cli/pkg/secret"
	"github.com/spf13/cobra"
	"os"
)

var getCmd = &cobra.Command{
	Use:   "get [args] [flags]",
	Short: "Returns the specified config format from a protected secret and generates credentials to 'current' location",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 2 {
			fmt.Printf("%s and %s is reqired arguments", SecretNameFlag, TeamFlag)
			os.Exit(1)
		}

		secretName := args[0]
		team := args[1]

		configType, err := helpers.GetString(cmd, ConfigFlag, false)
		if err != nil {
			fmt.Printf("getting %s: %s", ConfigFlag, err)
			os.Exit(1)
		}

		if configType != config.ENV && configType != config.ALL && configType != config.KCAT {
			fmt.Printf("valid args: %s | %s | %s", config.ENV, config.KCAT, config.ALL)
			os.Exit(1)
		}

		dest, err := helpers.GetString(cmd, DestFlag, false)
		if err != nil {
			fmt.Printf("getting %s: %s", DestFlag, err)
			os.Exit(1)
		}
		secret.ExtractAndGenerateConfig(configType, dest, secretName, team)
	},
}
