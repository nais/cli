package cli

import (
	"fmt"
	"strings"

	"github.com/nais/cli/internal/output"
	"github.com/spf13/cobra"
)

type Command struct {
	cobraCmd *cobra.Command
	output   output.Output

	validateFuncs []ValidateFunc
	subCommands   []*Command

	args []string
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
		PreRunE: func(co *cobra.Command, args []string) error {
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

func (c *Command) setup(w output.Output) {
	c.output = w
	for _, sub := range c.subCommands {
		sub.setup(w)
		c.cobraCmd.AddCommand(sub.cobraCmd)
	}
}
