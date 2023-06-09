package appstarter

import (
	"fmt"
	flags "github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/appstarter"
	"github.com/spf13/cobra"
)

func InitAppStarterCmd(rootCmd *cobra.Command) {
	appStarterCommand.Flags().StringP(flags.AppName, "n", "", "application name (as it will be in the cluster)")
	_ = appStarterCommand.MarkFlagRequired(flags.AppName)
	appStarterCommand.Flags().StringP(flags.TeamName, "t", "", "your team's name (app will be deployed to this namespace)")
	_ = appStarterCommand.MarkFlagRequired(flags.TeamName)
	appStarterCommand.Flags().StringSliceP(flags.Extras, "e", []string{}, "comma separated list of desired extras (idporten,openSearch,aad,postgres)")
	appStarterCommand.Flags().StringSliceP(flags.Topics, "c", []string{}, "comma separated list of desired kafka topic resources")
	appStarterCommand.Flags().UintP(flags.AppPortFlag, "p", 8080, "the port the app will listen on")
	rootCmd.AddCommand(appStarterCommand)
}

var appStarterCommand = &cobra.Command{
	Use:   "start [args]",
	Short: "Bootstrap basic yaml for nais and GitHub workflows",
	RunE: func(cmd *cobra.Command, args []string) error {
		appName, err := cmd.Flags().GetString(flags.AppName)
		if err != nil {
			return fmt.Errorf("error while collecting flag: %v", err)
		}
		teamName, err := cmd.Flags().GetString(flags.TeamName)
		if err != nil {
			return fmt.Errorf("error while collecting flag: %v", err)
		}
		extras, err := cmd.Flags().GetStringSlice(flags.Extras)
		if err != nil {
			return fmt.Errorf("error while collecting flag: %v", err)
		}
		topics, err := cmd.Flags().GetStringSlice(flags.Topics)
		if err != nil {
			return fmt.Errorf("error while collecting flag: %v", err)
		}
		appListenPort, err := cmd.Flags().GetUint(flags.AppPortFlag)
		if err != nil {
			return fmt.Errorf("error while collecting flag '%s': %v", flags.AppPortFlag, err)
		}
		return appstarter.Naisify(appName, teamName, extras, topics, appListenPort)
	},
}
