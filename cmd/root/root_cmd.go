package root

import (
	"fmt"
	"github.com/nais/nais-cli/cmd"
	"github.com/nais/nais-cli/cmd/root/aiven"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	VERSION string
	COMMIT  string
	DATE    string
	BuiltBy string

	rootCmd = &cobra.Command{
		Use:   "nais [command]",
		Short: "A simple NAIS CLI",
		Long:  `This is a NAIS tool to ease when working with NAIS clusters.`,
	}
)

func Execute(version, commit, date, builtBy string) {
	VERSION = version
	COMMIT = commit
	DATE = date
	BuiltBy = builtBy

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	aivenConfig := aiven.NewAivenConfig()
	aivenConfig.InitCmds(rootCmd)
	initVersionCmd()
}

func initConfig() {
	viper.SetEnvPrefix("NAIS_CLI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initVersionCmd() {
	versionCmd.Flags().BoolP(cmd.CommitInformation, "i", false, "Detailed commit information for this 'nais-cli' version (optional)")
	viper.BindPFlag(cmd.CommitInformation, versionCmd.Flags().Lookup(cmd.CommitInformation))
	rootCmd.AddCommand(versionCmd)
}
