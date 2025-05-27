package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/output"
	"github.com/spf13/cobra"
)

type (
	RunFunc      func(context.Context, output.Output, []string) error
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

func WithFlag[T flagTypes](name, short, usage string, value *T, opts ...FlagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, short, usage, value, c.cobraCmd.Flags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

func WithStickyFlag[T flagTypes](name, short, usage string, value *T, opts ...FlagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, short, usage, value, c.cobraCmd.PersistentFlags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

func WithRun(run RunFunc) CommandOption {
	return func(c *Command) {
		c.cobraCmd.RunE = func(co *cobra.Command, args []string) error {
			return run(co.Context(), c.output, args)
		}
	}
}

func WithValidate(validate ...ValidateFunc) CommandOption {
	return func(c *Command) {
		c.validateFuncs = append(c.validateFuncs, validate...)
	}
}

func WithAutoComplete(autocomplete func(ctx context.Context, args []string, toComplete string) ([]string, string)) CommandOption {
	return func(c *Command) {
		c.cobraCmd.ValidArgsFunction = func(co *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			suggestions, help := autocomplete(co.Context(), args, toComplete)
			if help != "" {
				suggestions = cobra.AppendActiveHelp(suggestions, help)
			}

			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

// WithAutoCompleteFiles sets up the command to suggest file completions with specific extensions.
func WithAutoCompleteFiles(ext ...string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.ValidArgsFunction = func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			helpSuffix := ""
			if num := len(ext); num > 0 {
				formatted := make([]string, num)
				for i, e := range ext {
					formatted[i] = "*." + e
				}
				helpSuffix = " (" + strings.Join(formatted[:num-1], ", ") + " or " + formatted[num-1] + ")"
			}

			ext = cobra.AppendActiveHelp(ext, fmt.Sprintf("Please choose one or more files%s.", helpSuffix))
			return ext, cobra.ShellCompDirectiveFilterFileExt
		}
	}
}
