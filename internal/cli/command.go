package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/version"
	"github.com/spf13/cobra"
)

type HandlerFunc func(context.Context) error

type Application struct {
	cobraCmd *cobra.Command

	Commands []*Command
}

type Command struct {
	cobraCmd *cobra.Command

	subCommands []*Command
}

func (a *Application) Run(ctx context.Context) error {
	a.setup()
	return a.cobraCmd.ExecuteContext(ctx)
}

func (a *Application) setup() {
	a.cobraCmd = &cobra.Command{
		Use:                "nais",
		Long:               "Nais CLI",
		Version:            version.Version + "-" + version.Commit,
		SilenceUsage:       true,
		DisableSuggestions: true,
	}

	for _, cmd := range a.Commands {
		cmd.setup()
		a.cobraCmd.AddCommand(cmd.cobraCmd)
	}
}

func NewCommand(name, short string, opts ...CommandOption) *Command {
	if strings.Contains(name, " ") {
		panic(fmt.Sprintf("command name cannot contain spaces: %v", name))
	}

	cmd := &Command{
		cobraCmd: &cobra.Command{
			Use:   name,
			Short: short,
		},
	}

	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}

func (c *Command) setup() {
	for _, sub := range c.subCommands {
		sub.setup()
		c.cobraCmd.AddCommand(sub.cobraCmd)
	}
}
