package root

import (
	"fmt"
	"github.com/nais/nais-cli/cmd"
	"github.com/nais/nais-cli/cmd/aiven"
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
	initTidyCmd()
}

func initConfig() {
	viper.SetEnvPrefix("NAIS_CLI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initAivenCmd() {
	rootCmd.AddCommand(aiven.AivenCommand)
}

func initVersionCmd() {
	VersionCmd.Flags().BoolP(cmd.CommitInformation, "i", false, "Detailed commit information for this 'nais-cli' version (optional)")
	viper.BindPFlag(cmd.CommitInformation, VersionCmd.Flags().Lookup(cmd.DestFlag))
	rootCmd.AddCommand(VersionCmd)
}

func initGetCmd() {
	aiven.GetCmd.Flags().StringP(cmd.DestFlag, "d", "", "If other then default 'tmp' folder (optional)")
	viper.BindPFlag(cmd.DestFlag, aiven.GetCmd.Flags().Lookup(cmd.DestFlag))

	aiven.GetCmd.Flags().StringP(cmd.ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(cmd.ConfigFlag, aiven.GetCmd.Flags().Lookup(cmd.ConfigFlag))

	aiven.AivenCommand.AddCommand(aiven.GetCmd)
}

func initCreateCmd() {
	aiven.CreateCmd.Flags().StringP(cmd.PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect (optional)")
	viper.BindPFlag(cmd.PoolFlag, aiven.CreateCmd.Flags().Lookup(cmd.PoolFlag))

	aiven.CreateCmd.Flags().IntP(cmd.ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(cmd.ExpireFlag, aiven.CreateCmd.Flags().Lookup(cmd.ExpireFlag))

	aiven.CreateCmd.Flags().StringP(cmd.SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(cmd.SecretNameFlag, aiven.CreateCmd.Flags().Lookup(cmd.SecretNameFlag))
	aiven.AivenCommand.AddCommand(aiven.CreateCmd)
}

func initTidyCmd() {
	aiven.TidyCmd.Flags().StringP(cmd.RootFlag, "r", "/var/", "temp folder other then '/var/' on Mac")
	viper.BindPFlag(cmd.RootFlag, aiven.TidyCmd.Flags().Lookup(cmd.RootFlag))
	aiven.AivenCommand.AddCommand(aiven.TidyCmd)
}
