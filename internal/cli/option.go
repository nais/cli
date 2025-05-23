package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CommandOption func(*Command)

func WithSubCommands(subCommands ...*Command) CommandOption {
	return func(c *Command) {
		c.subCommands = subCommands
	}
}

func WithPositionalArgs(args ...string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Use += " " + strings.ToUpper(strings.Join(args, " "))
	}
}

func WithLong(long string) CommandOption {
	return func(c *Command) {
		c.cobraCmd.Long = long
	}
}

func WithFlag[T flagTypes](name, usage, short string, value *T) CommandOption {
	return func(c *Command) {
		setupFlag(name, usage, short, value, c.cobraCmd.Flags())
	}
}

func WithStickyFlag[T flagTypes](name, usage, short string, value *T) CommandOption {
	return func(c *Command) {
		setupFlag(name, usage, short, value, c.cobraCmd.PersistentFlags())
	}
}

func WithHandler(handler HandlerFunc) CommandOption {
	return func(c *Command) {
		c.cobraCmd.RunE = func(co *cobra.Command, args []string) error {
			return handler(co.Context())
		}
	}
}

func setupFlag(name, usage, short string, value any, flags *pflag.FlagSet) {
	switch ptr := value.(type) {
	case *string:
		if short == "" {
			flags.StringVar(ptr, name, "", usage)
		} else {
			flags.StringVarP(ptr, name, short, "", usage)
		}
	case *bool:
		if short == "" {
			flags.BoolVar(ptr, name, false, usage)
		} else {
			flags.BoolVarP(ptr, name, short, false, usage)
		}
	case *int:
		if short == "" {
			flags.CountVar(ptr, name, usage)
		} else {
			flags.CountVarP(ptr, name, short, usage)
		}
	default:
		panic(fmt.Sprintf("unknown flag type: %T", value))
	}
}
