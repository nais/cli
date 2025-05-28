package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/output"
	"github.com/spf13/cobra"
)

type (
	// RunFunc is a function that will be executed when the command is run.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order.
	RunFunc func(ctx context.Context, out output.Output, args []string) error

	// ValidateFunc is a function that will be executed before the command's RunFunc is executed.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order.
	ValidateFunc func(ctx context.Context, args []string) error

	// AutoCompleteFunc is a function that will be executed to provide auto-completion suggestions for the command.
	//
	// The args passed to this function is the arguments added to the command using the WithArgs option, in the same
	// order. toComplete is the current input that the user is typing, and it can be used to filter the suggestions.
	// The first return value is a slice of strings that will be used as suggestions, and the second return value is a
	// string that will be used as active help text in the shell while performing auto-complete.
	AutoCompleteFunc func(ctx context.Context, args []string, toComplete string) (completions []string, activeHelp string)
)

// CommandOption defines a function that modifies a Command instance.
type CommandOption func(*Command)

// WithSubCommands adds subcommands to the command.
func WithSubCommands(subCommands ...*Command) CommandOption {
	return func(c *Command) {
		c.subCommands = subCommands
	}
}

// WithArgs adds positional arguments to the command. The arguments will be injected into the commands RunFunc.
func WithArgs(args ...string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Use += " " + strings.ToUpper(strings.Join(args, " "))
	}
}

// WithLongDescription adds a long description to the command used for help output.
func WithLongDescription(desc string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Long = desc
	}
}

// WithFlag sets up a flag for the command. Use FlagOption to customize the flag further.
func WithFlag[T flagTypes](name, short, usage string, value *T, opts ...FlagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, short, usage, value, c.cobraCmd.Flags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

// WithStickyFlag sets up a flag that is persistent across all subcommands. Use FlagOption to customize the flag
// further.
func WithStickyFlag[T flagTypes](name, short, usage string, value *T, opts ...FlagOption) CommandOption {
	return func(c *Command) {
		setupFlag(name, short, usage, value, c.cobraCmd.PersistentFlags())
		for _, opt := range opts {
			opt(c.cobraCmd, name)
		}
	}
}

// InGroup places the command in a specific group. This is mainly used for grouping of commands in the help text.
func InGroup(group string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.GroupID = group
	}
}

// WithRun sets up the command handler function that will be executed when the command is run.
func WithRun(f RunFunc) CommandOption {
	return func(c *Command) {
		c.cobraCmd.RunE = func(co *cobra.Command, args []string) error {
			return f(co.Context(), c.output, args)
		}
	}
}

// WithValidate adds validation functions that will be executed before the command's RunFunc is executed. The validation
// functions will be executed in the added order, and if one of them returns an error the RunFunc will not be executed.
func WithValidate(f ...ValidateFunc) CommandOption {
	return func(c *Command) {
		c.validateFuncs = append(c.validateFuncs, f...)
	}
}

// WithAutoComplete sets up a function that will be used to provide auto-completion suggestions for the command.
func WithAutoComplete(f AutoCompleteFunc) CommandOption {
	return func(c *Command) {
		c.cobraCmd.ValidArgsFunction = func(co *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			suggestions, help := f(co.Context(), args, toComplete)
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
