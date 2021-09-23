package root

import (
	"fmt"
	"github.com/nais/nais-cli/cmd/aiven"
	"github.com/nais/nais-cli/cmd/helpers"
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
	rootCmd.AddCommand(aiven.AivenCommand)
}

func initVersionCmd() {
	VersionCmd.Flags().BoolP(helpers.CommitInformation, "i", false, "Detailed commit information for this 'nais-cli' version (optional)")
	viper.BindPFlag(helpers.CommitInformation, VersionCmd.Flags().Lookup(helpers.DestFlag))
	rootCmd.AddCommand(VersionCmd)
}

func initGetCmd() {
	aiven.GetCmd.Flags().StringP(helpers.DestFlag, "d", "", "Path to directory where secrets will be dropped of. For current './creds' (optional)")
	viper.BindPFlag(helpers.DestFlag, aiven.GetCmd.Flags().Lookup(helpers.DestFlag))

	aiven.GetCmd.Flags().StringP(helpers.ConfigFlag, "c", "all", "Type of config do be generated, supported ( .env || kcat || all ) (optional)")
	viper.BindPFlag(helpers.ConfigFlag, aiven.GetCmd.Flags().Lookup(helpers.ConfigFlag))

	aiven.AivenCommand.AddCommand(aiven.GetCmd)
}

func initCreateCmd() {
	aiven.CreateCmd.Flags().StringP(helpers.PoolFlag, "p", "nav-dev", "Preferred kafka pool to connect (optional)")
	viper.BindPFlag(helpers.PoolFlag, aiven.CreateCmd.Flags().Lookup(helpers.PoolFlag))

	aiven.CreateCmd.Flags().IntP(helpers.ExpireFlag, "e", 1, "Time in days the created secret should be valid (optional)")
	viper.BindPFlag(helpers.ExpireFlag, aiven.CreateCmd.Flags().Lookup(helpers.ExpireFlag))

	aiven.CreateCmd.Flags().StringP(helpers.SecretNameFlag, "s", "", "Preferred secret-name instead of generated (optional)")
	viper.BindPFlag(helpers.SecretNameFlag, aiven.CreateCmd.Flags().Lookup(helpers.SecretNameFlag))
	aiven.AivenCommand.AddCommand(aiven.CreateCmd)
}
