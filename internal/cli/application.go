package cli

import (
	"context"
	"iter"
	"maps"
	"os"
	"slices"

	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

const (
	GroupAuthentication = "Authentication"
)

type Application struct {
	cobraCmd *cobra.Command
}

func NewApplication(flags any, cmd ...*Command) *Application {
	cobra.EnableTraverseRunHooks = true

	cc := &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}

	setupFlags(flags, cc.PersistentFlags())

	w := output.NewWriter(cc.OutOrStdout())

	for group := range allGroups(cmd) {
		cc.AddGroup(&cobra.Group{
			ID:    group,
			Title: group,
		})
	}

	for _, c := range cmd {
		c.init(w)
		cc.AddCommand(c.cobraCmd)
	}

	return &Application{
		cobraCmd: cc,
	}
}

type LogLevelFlags interface {
	IsVerbose() bool
}

func (a *Application) Run(ctx context.Context, flags LogLevelFlags) error {
	autoComplete := slices.Contains(os.Args[1:], "__complete")

	if !autoComplete {
		flushMetrics := metric.Initialize()
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
			flushMetrics(flags.IsVerbose())
		}()
	}

	executedCommand, err := a.cobraCmd.ExecuteContextC(ctx)
	if !autoComplete && executedCommand != nil {
		collectCommandHistogram(ctx, executedCommand, err)
	}

	if err != nil {
		return err
	}

	return nil
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
