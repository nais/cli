package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	VERSION string

	rootCmd = &cobra.Command{
		Use:   "debuk [COMMANDS] [FLAGS]",
		Short: "A generator for AivenApplications",
		Long: `Debuk is a CLI. 
This application is a tool to generate the needed files to quickly start debugging your aivenApplication topics.`,
	}
)

func Execute(version string) {
	VERSION = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const (
	TeamFlag       = "team"
	UsernameFlag   = "username"
	DestFlag       = "dest"
	ExpireFlag     = "expire"
	PoolFlag       = "pool"
	SecretNameFlag = "secret-name"
)

func init() {
	cobra.OnInitialize(initConfig)

	applyCommand.Flags().StringP(UsernameFlag, "u", "", "Username for the aivenApplication configuration (required)")
	viper.BindPFlag(UsernameFlag, applyCommand.Flags().Lookup(UsernameFlag))

	applyCommand.Flags().StringP(TeamFlag, "t", "", "Teamnamespace that the user have access to (required)")
	viper.BindPFlag(TeamFlag, applyCommand.Flags().Lookup(TeamFlag))

	applyCommand.Flags().StringP(PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect (optional)")
	viper.BindPFlag(PoolFlag, applyCommand.Flags().Lookup(PoolFlag))

	applyCommand.Flags().IntP(ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(ExpireFlag, applyCommand.Flags().Lookup(ExpireFlag))

	applyCommand.Flags().StringP(DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(DestFlag, applyCommand.Flags().Lookup(DestFlag))

	applyCommand.Flags().StringP(SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(SecretNameFlag, applyCommand.Flags().Lookup(SecretNameFlag))

	rootCmd.AddCommand(applyCommand)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	viper.SetEnvPrefix("DEBUK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}
