package cli

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nais/cli/internal/output"
	"github.com/spf13/cobra"
)

type Command struct {
	cobraCmd *cobra.Command
	output   output.Output

	validateFuncs []ValidateFunc
	subCommands   []*Command
}

func (c *Command) setupFlags(flags any, level int) {
	fields := reflect.TypeOf(flags).Elem()
	values := reflect.ValueOf(flags).Elem()

	for i := range fields.NumField() {
		field := fields.Field(i)
		value := values.Field(i)

		if !field.IsExported() {
			fmt.Printf("skipping: unexported field %v %v\n", field.Name, value.String())
			continue
		} else {
			fmt.Printf("processing field: %v %v\n", field.Name, value.String())
		}

		// or is it just optional?
		if value.Kind() == reflect.Pointer && value.Elem().Kind() == reflect.Struct {
			c.setupFlags(value.Interface(), level+1)
		} else {
			flagName, ok := field.Tag.Lookup("name")
			if !ok {
				flagName = strings.ToLower(field.Name)
			}

			flagUsage, ok := field.Tag.Lookup("usage")
			if !ok {
				flagUsage = field.Name
			}
			flagShort, ok := field.Tag.Lookup("short")
			if !ok {
				flagShort = ""
			}

			if !value.CanAddr() {
				panic(fmt.Sprintf("field %v is not addressable, cannot set up flag", field.Name))
			}

			if level == 0 {
				setupFlag(flagName, flagShort, flagUsage, value.Addr().Interface(), c.cobraCmd.Flags())
			} else {
				setupFlag(flagName, flagShort, flagUsage, value.Addr().Interface(), c.cobraCmd.PersistentFlags())
			}
		}
	}
}

func NewCommand(name, short string, opts ...CommandOption) *Command {
	if strings.Contains(name, " ") {
		panic(fmt.Sprintf("command name cannot contain spaces: %v", name))
	}

	cmd := &Command{}
	cmd.cobraCmd = &cobra.Command{
		Use:   name,
		Short: short,
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		PersistentPreRunE: func(co *cobra.Command, args []string) error {
			for _, validate := range cmd.validateFuncs {
				if err := validate(co.Context(), args); err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}
			}
			return nil
		},
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func (c *Command) setup(out output.Output) {
	c.output = out
	for _, sub := range c.subCommands {
		sub.setup(out)
		c.cobraCmd.AddCommand(sub.cobraCmd)
	}
}
