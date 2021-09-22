package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	VERSION  string
	COMMIT   string
	DATE     string
	BUILT_BY string

	rootCmd = &cobra.Command{
		Use:   "nais [command] [args] [flags]",
		Short: "A simple NAIS client to generate resources for debug purpose",
		Long: `NAIS debug CLI. 
This is a NAIS tool to extract secrets from cluster to quickly start debugging your NAIS resources.`,
	}
)

func Execute(version, commit, date, builtBy string) {
	VERSION = version
	COMMIT = commit
	DATE = date
	BUILT_BY = builtBy

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	initAivenCmd()
	initVersionCmd()
	initGetCmd()
	initCreateCmd()
}

func initConfig() {
	viper.SetEnvPrefix("NAIS_CLI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initAivenCmd() {
	rootCmd.AddCommand(AivenCommand)
}

func initVersionCmd() {
	versionCmd.Flags().BoolP(CommitInformation, "i", false, "Detailed commit information for this 'nais-cli' version (optional)")
	viper.BindPFlag(CommitInformation, versionCmd.Flags().Lookup(DestFlag))
	rootCmd.AddCommand(versionCmd)
}

func initGetCmd() {
	GetCmd.Flags().StringP(DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(DestFlag, GetCmd.Flags().Lookup(DestFlag))

	GetCmd.Flags().StringP(ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(ConfigFlag, GetCmd.Flags().Lookup(ConfigFlag))

	AivenCommand.AddCommand(GetCmd)
}

func initCreateCmd() {
	CreateCmd.Flags().StringP(PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect (optional)")
	viper.BindPFlag(PoolFlag, CreateCmd.Flags().Lookup(PoolFlag))

	CreateCmd.Flags().IntP(ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(ExpireFlag, CreateCmd.Flags().Lookup(ExpireFlag))

	CreateCmd.Flags().StringP(SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(SecretNameFlag, CreateCmd.Flags().Lookup(SecretNameFlag))
	AivenCommand.AddCommand(CreateCmd)
}
