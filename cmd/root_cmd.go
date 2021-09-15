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
		Use:   "nais-d [command] [args] [flags]",
		Short: "A simple nais client to generate resources for debug",
		Long: `nais-d debug CLI. 
This is a nais-debug-tool to extract secrets from cluster to quickly start debugging your nais resources.`,
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
	getCmd.Flags().StringP(DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(DestFlag, getCmd.Flags().Lookup(DestFlag))

	getCmd.Flags().StringP(ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(ConfigFlag, getCmd.Flags().Lookup(ConfigFlag))

	aivenCommand.AddCommand(getCmd)
}
