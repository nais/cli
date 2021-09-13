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
		Use:   "nais-d [COMMANDS] [FLAGS]",
		Short: "A simple nais client to generate resources for debug",
		Long: `nais-d is a CLI. 
This application is a tool to generate the needed files to quickly start debugging your nais resources.`,
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

const (
	TeamFlag       = "team"
	UsernameFlag   = "username"
	DestFlag       = "dest"
	ConfigFlag     = "config"
	ExpireFlag     = "expire"
	PoolFlag       = "pool"
	SecretNameFlag = "secret-name"

	CommitInformation = "commit"
)

func init() {
	cobra.OnInitialize(initConfig)
	initAivenCmd()
	initVersionCmd()
	initGetCmd()
}

func initConfig() {
	viper.SetEnvPrefix("NAIS_D")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initAivenCmd() {
	aivenCommand.Flags().StringP(UsernameFlag, "u", "", "Username for the aivenApplication configuration (required)")
	viper.BindPFlag(UsernameFlag, aivenCommand.Flags().Lookup(UsernameFlag))

	aivenCommand.Flags().StringP(TeamFlag, "t", "", "Team-namespace that the user have access to (required)")
	viper.BindPFlag(TeamFlag, aivenCommand.Flags().Lookup(TeamFlag))

	aivenCommand.Flags().StringP(PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect (optional)")
	viper.BindPFlag(PoolFlag, aivenCommand.Flags().Lookup(PoolFlag))

	aivenCommand.Flags().IntP(ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(ExpireFlag, aivenCommand.Flags().Lookup(ExpireFlag))

	aivenCommand.Flags().StringP(SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(SecretNameFlag, aivenCommand.Flags().Lookup(SecretNameFlag))
	rootCmd.AddCommand(aivenCommand)
}

func initVersionCmd() {
	versionCmd.Flags().BoolP(CommitInformation, "i", false, "Detailed commit information for this 'nais-d' version (optional)")
	viper.BindPFlag(CommitInformation, versionCmd.Flags().Lookup(DestFlag))
	rootCmd.AddCommand(versionCmd)
}

func initGetCmd() {
	getCmd.Flags().StringP(SecretNameFlag, "s", "", "Secret-name specified for aiven application (required)")
	viper.BindPFlag(SecretNameFlag, getCmd.Flags().Lookup(SecretNameFlag))

	getCmd.Flags().StringP(TeamFlag, "t", "", "Team-namespace that the user have access to (required)")
	viper.BindPFlag(TeamFlag, getCmd.Flags().Lookup(TeamFlag))

	getCmd.Flags().StringP(DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(DestFlag, getCmd.Flags().Lookup(DestFlag))

	getCmd.Flags().StringP(ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(ConfigFlag, getCmd.Flags().Lookup(ConfigFlag))

	aivenCommand.AddCommand(getCmd)
}
