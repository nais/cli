package cli

import (
	"context"
	"iter"
	"maps"

	"github.com/spf13/cobra"
)

type Application struct {
	// Name is the name of the application, used as the root command in the CLI.
	Name string

	// Title is the title of the application, used as a short description for the help output.
	Title string

	// Version is the version of the application, used in the help output.
	Version string

	// StickyFlags are flags that should be available for all subcommands of the application.
	StickyFlags any

	// SubCommands are the commands that are part of the application.
	SubCommands []*Command

	cobraCmd *cobra.Command
}

func (a *Application) Run(ctx context.Context, out Output, args []string) ([]string, error) {
	cobra.EnableTraverseRunHooks = true

	a.cobraCmd = &cobra.Command{
		Use:                a.Name,
		Short:              a.Title,
		Version:            a.Version,
		SilenceErrors:      true,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}
	a.cobraCmd.SetArgs(args)
	a.cobraCmd.SetOut(out)

	setupFlags(a.cobraCmd, a.StickyFlags, a.cobraCmd.PersistentFlags())

	for group := range allGroups(a.SubCommands) {
		a.cobraCmd.AddGroup(&cobra.Group{
			ID:    group,
			Title: group,
		})
	}

	for _, c := range a.SubCommands {
		c.init(a.Name, out)
		a.cobraCmd.AddCommand(c.cobraCmd)
	}

	executedCommand, err := a.cobraCmd.ExecuteContextC(ctx)
	return commandNames(executedCommand), err
}

func allGroups(cmds []*Command) iter.Seq[string] {
	var rec func(cmds []*Command, groups map[string]struct{})
	rec = func(cmds []*Command, groups map[string]struct{}) {
		for _, cmd := range cmds {
			if cmd.Group != "" {
				groups[cmd.Group] = struct{}{}
			}
			rec(cmd.SubCommands, groups)
		}
	}

	groups := make(map[string]struct{})
	rec(cmds, groups)

	return maps.Keys(groups)
}

func commandNames(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}

	return append(commandNames(cmd.Parent()), cmd.Name())
}
