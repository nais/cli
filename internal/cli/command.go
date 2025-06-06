package cli

import (
	"fmt"
	"strings"

	"github.com/nais/cli/internal/output"
	"github.com/spf13/cobra"
)

type Argument struct {
	// Name is the name of the argument, used for help output.
	Name string

	// Required can be set if the argument is required.
	Required bool

	// Repeatable can be used for repeatable arguments, this can only be set for the last argument in the command.
	Repeatable bool
}

type Command struct {
	// Name is the name of the command, this is used to invoke the command in the CLI. This field is required.
	Name string

	// Short is the short description for the command, shown in the help output. This field is required.
	Short string

	// Long is the long description for the command, shown in the help output.
	Long string

	// RunFunc will be executed when the command is run.
	RunFunc RunFunc

	// ValidateFunc will be executed before the command's RunFunc is executed.
	ValidateFunc ValidateFunc

	// AutoCompleteFunc sets up a function that will be used to provide auto-completion suggestions for the command.
	AutoCompleteFunc AutoCompleteFunc

	// AutoCompleteExtensions specifies which file extensions to list in autocompletion. This overrides
	// AutoCompleteFunc.
	AutoCompleteExtensions []string

	// Group places the command in a specific group. This is mainly used for grouping of commands in the help text.
	Group string

	// SubCommands adds subcommands to the command.
	SubCommands []*Command

	// Args are the positional arguments to the command. The arguments will be injected into the commands RunFunc.
	Args []Argument

	// Flags sets up flags for the command.
	Flags any

	// StickyFlags sets up flags that is persistent across all subcommands.
	StickyFlags any

	cobraCmd *cobra.Command
}

// use generates a normalized use string for the internal Cobra command.
func use(cmd string, args []Argument) string {
	for i, arg := range args {
		if arg.Name == "" {
			panic(fmt.Sprintf("argument name (%+v) cannot be empty", arg))
		}

		suffix := ""
		if arg.Repeatable {
			if i != len(args)-1 {
				panic(fmt.Sprintf("repeatable argument (%+v) must be the last argument in the command", arg))
			}
			suffix = "..."
		}

		format := " %s%s"
		if !arg.Required {
			format = " [%s%s]"
		}

		cmd += fmt.Sprintf(format, strings.ToUpper(arg.Name), suffix)
	}

	return cmd
}

func run(f RunFunc, out output.Output) func(*cobra.Command, []string) error {
	if f != nil {
		return func(co *cobra.Command, args []string) error {
			return f(co.Context(), out, args)
		}
	}
	return nil
}

func (c *Command) init(out output.Output) {
	if strings.Contains(c.Name, " ") {
		panic(fmt.Sprintf("command name cannot contain spaces: %v", c.Name))
	}

	c.cobraCmd = &cobra.Command{
		Use:               use(c.Name, c.Args),
		Short:             c.Short,
		Long:              c.Long,
		GroupID:           c.Group,
		RunE:              run(c.RunFunc, out),
		ValidArgsFunction: autocomplete(c.AutoCompleteFunc, c.AutoCompleteExtensions),
		PersistentPreRunE: func(co *cobra.Command, args []string) error {
			if c.ValidateFunc == nil {
				return nil
			}

			if err := c.ValidateFunc(co.Context(), args); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			return nil
		},
	}

	setupFlags(c.Flags, c.cobraCmd.Flags())
	setupFlags(c.StickyFlags, c.cobraCmd.PersistentFlags())

	for _, sub := range c.SubCommands {
		sub.init(out)
		c.cobraCmd.AddCommand(sub.cobraCmd)
	}
}
