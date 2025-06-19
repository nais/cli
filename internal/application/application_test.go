package application

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/cli/internal/root"
)

func TestHelpForAllCommands(t *testing.T) {
	ctx := context.Background()
	app := newApplication(&root.Flags{})
	for _, cmd := range app.SubCommands {
		t.Run("Generate help "+cmd.Name, func(t *testing.T) {
			runCommand(t, ctx, cmd, []string{})
		})
	}
}

func runCommand(t *testing.T, ctx context.Context, cmd *cli.Command, parentCommands []string) {
	t.Helper()

	args := append(parentCommands, cmd.Name)

	out := cli.NewWriter(&bytes.Buffer{})
	helpCmd := append(args, "--help")

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("failed to run command %q: %v", strings.Join(helpCmd, " "), err)
		}
	}()
	_, err := newApplication(&root.Flags{}).Run(ctx, out, helpCmd)
	if err != nil {
		t.Fatalf("failed to run command %s: %v", strings.Join(helpCmd, " "), err)
	}

	for _, subCmd := range cmd.SubCommands {
		t.Run(subCmd.Name, func(t *testing.T) {
			runCommand(t, ctx, subCmd, args)
		})
	}
}
