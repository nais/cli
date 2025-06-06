package cli

import (
	"context"
	"iter"
	"maps"

	"github.com/spf13/cobra"
)

type Application struct {
	Name        string
	Long        string
	Version     string
	StickyFlags any
	SubCommands []*Command
	cobraCmd    *cobra.Command
}

func (a *Application) Run(ctx context.Context, out Output) ([]string, error) {
	cobra.EnableTraverseRunHooks = true

	a.cobraCmd = &cobra.Command{
		Use:                a.Name,
		Long:               a.Long,
		Version:            a.Version,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}

	setupFlags(a.StickyFlags, a.cobraCmd.PersistentFlags())

	for group := range allGroups(a.SubCommands) {
		a.cobraCmd.AddGroup(&cobra.Group{
			ID:    group,
			Title: group,
		})
	}

	for _, c := range a.SubCommands {
		c.init(out)
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
