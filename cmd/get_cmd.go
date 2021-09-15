package cmd

import (
	"fmt"
	"github.com/nais/nais-d/cmd/helpers"
	"github.com/nais/nais-d/pkg/consts"
	"github.com/nais/nais-d/pkg/secret"
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

		if configType != consts.ENV && configType != consts.ALL && configType != consts.KCAT {
			fmt.Printf("valid args: %s | %s | %s", consts.ENV, consts.KCAT, consts.ALL)
			os.Exit(1)
		}

		dest, err := helpers.GetString(cmd, DestFlag, false)
		if err != nil {
			fmt.Printf("getting %s: %s", DestFlag, err)
			os.Exit(1)
		}

		dest, err = helpers.DefaultDestination(dest)
		if err != nil {
			fmt.Printf("an error %s", err)
			os.Exit(1)
		}
		secret.ExtractAndGenerateConfig(configType, dest, secretName, team)
	},
}
