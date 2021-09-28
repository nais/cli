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
	VERSION string
	COMMIT  string
	DATE    string
	BuiltBy string

	RootCmd = &cobra.Command{
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

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	aivenConfig := NewAivenConfig(
		aiven.AivenCommand,
		aiven.CreateCmd,
		aiven.GetCmd,
		aiven.TidyCmd,
	)
	aivenConfig.InitCmds()
	initVersionCmd()
}

func initConfig() {
	viper.SetEnvPrefix("NAIS_CLI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
}

func initVersionCmd() {
	VersionCmd.Flags().BoolP(cmd.CommitInformation, "i", false, "Detailed commit information for this 'nais-cli' version (optional)")
	viper.BindPFlag(cmd.CommitInformation, VersionCmd.Flags().Lookup(cmd.DestFlag))
	RootCmd.AddCommand(VersionCmd)
}
