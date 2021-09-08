package cmd

import (
	"fmt"
	"github.com/nais/debuk/pkg/aiven/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	KafkaNavDev  = "nav-dev"
	KafkaNavProd = "nav-prod"
	KafkaNavIntegrationTest = "nav-integration-test"
)

var applyCommand = &cobra.Command{
	Use:   "apply [FLAGS]",
	Short: "Apply a aivenApplication and fetch secrets",
	Long:  `Will apply a aivenApplication based on information given and extract credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {

		username, err := getString(cmd, UsernameFlag, true)
		if err != nil {
			return err
		}

		team, err := getString(cmd, TeamFlag, true)
		if err != nil {
			return err
		}

		pool, _ := getString(cmd, PoolFlag, false)
		if pool != KafkaNavDev && pool != KafkaNavProd && pool != KafkaNavIntegrationTest  {
			return fmt.Errorf("valid values for '--%s': %s | %s | %s", PoolFlag, KafkaNavDev, KafkaNavProd, KafkaNavIntegrationTest)
		}

		dest, err := getString(cmd, DestFlag, false)
		if err != nil {
			return fmt.Errorf("getting %s: %s", DestFlag, err)
		}

		dest, err = DefaultDestination(dest)
		if err != nil {
			return fmt.Errorf("setting destination: %s", err)
		}

		expiry, err := cmd.Flags().GetInt(ExpireFlag)
		secretName, err := getString(cmd, SecretNameFlag, false)
		if err != nil {
			return fmt.Errorf("getting flag %s", err)
		}

		if err := generate.AivenApplication(username, team, pool, dest, expiry, secretName); err != nil {
			return fmt.Errorf("an error occurred generating aivenApplication: %s", err)
		}
		return nil
	},
}

func getString(cmd *cobra.Command, flag string, required bool) (string, error) {
	env := viper.GetString(flag)
	if env != "" {
		return env, nil
	}
	arg, err := cmd.Flags().GetString(flag)
	if err != nil {
		return "", fmt.Errorf("getting %s: %s", flag, err)
	}
	if arg == "" {
		if required {
			return "", fmt.Errorf("%s is reqired", flag)
		}
	}
	return arg, nil

}
