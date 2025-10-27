package application

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nais/naistrix"
)

func TestHelpForAllCommands(t *testing.T) {
	ctx := context.Background()
	app, _, err := newApplication(io.Discard)
	if err != nil {
		t.Fatalf("unable to create application: %v", err)
	}

	for _, cmd := range app.Commands {
		t.Run("Generate help "+cmd.Name, func(t *testing.T) {
			runCommand(t, ctx, app.Application, cmd, []string{})
		})
	}
}

func runCommand(t *testing.T, ctx context.Context, app *naistrix.Application, cmd *naistrix.Command, parentCommands []string) {
	t.Helper()

	args := append(parentCommands, cmd.Name)
	helpCmd := append(args, "--help")

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("failed to run command %q: %v", strings.Join(helpCmd, " "), err)
		}
	}()
	err := app.Run(
		naistrix.RunWithContext(ctx),
		naistrix.RunWithArgs(helpCmd),
	)
	if err != nil {
		t.Fatalf("failed to run command %s: %v", strings.Join(helpCmd, " "), err)
	}

	for _, subCmd := range cmd.SubCommands {
		t.Run(subCmd.Name, func(t *testing.T) {
			runCommand(t, ctx, app, subCmd, args)
		})
	}
}
