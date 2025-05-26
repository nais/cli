package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

type (
	RunFunc      func(context.Context, []string) error
	ValidateFunc func(context.Context, []string) error
)

type CommandOption func(*Command)

func WithSubCommands(subCommands ...*Command) CommandOption {
	return func(c *Command) {
		c.subCommands = subCommands
	}
}

func WithArgs(args ...string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Use += " " + strings.ToUpper(strings.Join(args, " "))
	}
}

func WithLong(long string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Long = long
	}
}

func WithFlag[T flagTypes](name, usage, short string, value *T, opts ...flagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, usage, short, value, c.cobraCmd.Flags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

func WithStickyFlag[T flagTypes](name, usage, short string, value *T, opts ...flagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, usage, short, value, c.cobraCmd.PersistentFlags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

func WithRun(run RunFunc) CommandOption {
	return func(c *Command) {
		c.cobraCmd.RunE = func(co *cobra.Command, args []string) error {
			return run(co.Context(), args)
		}
	}
}

func WithValidate(validate ValidateFunc) CommandOption {
	return func(c *Command) {
		c.validateFuncs = append(c.validateFuncs, validate)
	}
}

func WithAutoComplete(autocomplete func(ctx context.Context, args []string) ([]string, string)) CommandOption {
	return func(c *Command) {
		c.cobraCmd.ValidArgsFunction = func(co *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			suggestions, help := autocomplete(co.Context(), args)
			if help != "" {
				suggestions = cobra.AppendActiveHelp(suggestions, help)
			}
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}
	}
}
