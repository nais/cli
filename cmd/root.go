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
		Use:   "debuk [COMMANDS] [FLAGS]",
		Short: "A generator for AivenApplications",
		Long: `Debuk is a CLI. 
This application is a tool to generate the needed files to quickly start debugging your aivenApplication topics.`,
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
	initApplyCmd()
	initVersionCmd()
	initGetCmd()
}

func initConfig() {
	viper.SetEnvPrefix("DEBUK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initApplyCmd() {
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
}

func initVersionCmd() {
	versionCmd.Flags().BoolP(CommitInformation, "i", false, "Detailed commit information for this debuk version (optional)")
	viper.BindPFlag(CommitInformation, versionCmd.Flags().Lookup(DestFlag))
	rootCmd.AddCommand(versionCmd)
}

func initGetCmd() {
	getCmd.Flags().StringP(DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(DestFlag, getCmd.Flags().Lookup(DestFlag))

	getCmd.Flags().StringP(ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(ConfigFlag, getCmd.Flags().Lookup(ConfigFlag))

	getCmd.Flags().StringP(SecretNameFlag, "s", "", "Secretname specified for aiven application (required)")
	viper.BindPFlag(SecretNameFlag, getCmd.Flags().Lookup(SecretNameFlag))
	rootCmd.AddCommand(getCmd)
}
