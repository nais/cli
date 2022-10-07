package root

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nais/cli/cmd/root/appstarter"
	"github.com/nais/cli/cmd/root/validate"

	"github.com/nais/cli/cmd"
	"github.com/nais/cli/cmd/root/aiven"
	"github.com/nais/cli/cmd/root/appstarter"
	"github.com/nais/cli/cmd/root/device"
	"github.com/nais/cli/cmd/root/naas"
	"github.com/nais/cli/cmd/root/postgres"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	VERSION string
	COMMIT  string
	DATE    string
	BuiltBy string

	rootCmd = &cobra.Command{
		Use:           "nais [command]",
		Short:         "A simple NAIS CLI",
		Long:          `This is a NAIS tool to ease when working with NAIS clusters.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute(version, commit, date, builtBy string) {
	VERSION = version
	COMMIT = commit
	DATE = date
	BuiltBy = builtBy

	var ctx context.Context

	for _, arg := range os.Args {
		// Since there's a default timeout of 10 seconds, we need to disable it for the "naas" and "postgres" subcommand.
		if arg == "naas" || arg == "postgres" {
			ctx = context.Background()
			break
		}
	}

	if ctx == nil {
		const timeout = 10 * time.Second
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	aivenConfig := aiven.NewConfig()
	aivenConfig.InitCmds(rootCmd)
	deviceConfig := device.NewDeviceConfig()
	deviceConfig.InitCmds(rootCmd)
	postgresConfig := postgres.NewConfig()
	postgresConfig.InitCmds(rootCmd)
	naasConfig := naas.NewConfig()
	naasConfig.InitCmds(rootCmd)
	initVersionCmd()
	appstarter.InitAppStarterCmd(rootCmd)
	validate.InitValidateCmd(rootCmd)
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
