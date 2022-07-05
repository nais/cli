package appstarter

import (
	"fmt"
	flags "github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/appstarter"
	"github.com/spf13/cobra"
)

func InitAppStarterCmd(rootCmd *cobra.Command) {
	appStarterCommand.Flags().StringP(flags.AppName, "n", "", "usagestuff appname goes here")
	_ = appStarterCommand.MarkFlagRequired(flags.AppName)
	appStarterCommand.Flags().StringP(flags.TeamName, "t", "", "usagestuff teamname goes here")
	_ = appStarterCommand.MarkFlagRequired(flags.TeamName)
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
		return appstarter.Naisify(appName, teamName, []string{}, []string{})
	},
}
