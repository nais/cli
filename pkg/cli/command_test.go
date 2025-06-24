package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nais/cli/pkg/cli"
)

var emptyRun = func(context.Context, cli.Output, []string) error {
	return nil
}

func TestCommandValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("command with no name", func(t *testing.T) {
		app := &cli.Application{
			Name: "app",
			SubCommands: []*cli.Command{
				{
					Title:   "Test command",
					RunFunc: emptyRun,
				},
			},
		}

		defer func() {
			contains := "cannot be empty"
			if r := recover(); r == nil {
				t.Fatalf("expected panic for command with no name, but did not panic")
			} else if msg := r.(string); !strings.Contains(msg, contains) {
				t.Fatalf("expected panic message to contain %q, got: %q", contains, msg)
			}
		}()
		_, _ = app.Run(ctx, cli.Stdout(), []string{"-h"})
	})

	t.Run("command with space in name", func(t *testing.T) {
		app := &cli.Application{
			Name: "app",
			SubCommands: []*cli.Command{
				{
					Name:    "test command",
					Title:   "Test command",
					RunFunc: emptyRun,
				},
			},
		}

		defer func() {
			contains := "cannot contain spaces: test command"
			if r := recover(); r == nil {
				t.Fatalf("expected panic for command with no name, but did not panic")
			} else if msg := r.(string); !strings.Contains(msg, contains) {
				t.Fatalf("expected panic message to contain %q, got: %q", contains, msg)
			}
		}()
		_, _ = app.Run(ctx, cli.Stdout(), []string{"-h"})
	})
}

func TestUseString(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		expectedArgsString string
		args               []cli.Argument
	}{
		{
			name:               "no arguments",
			expectedArgsString: "",
		},
		{
			name:               "optional argument",
			expectedArgsString: "[ARG]",
			args: []cli.Argument{
				{Name: "arg"},
			},
		},
		{
			name:               "required argument",
			expectedArgsString: "ARG",
			args: []cli.Argument{
				{Name: "arg", Required: true},
			},
		},
		{
			name:               "optional repeatable argument",
			expectedArgsString: "[ARG...]",
			args: []cli.Argument{
				{Name: "arg", Repeatable: true},
			},
		},
		{
			name:               "required repeatable argument",
			expectedArgsString: "ARG [ARG...]",
			args: []cli.Argument{
				{Name: "arg", Required: true, Repeatable: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Application{
				Name: "app",
				SubCommands: []*cli.Command{
					{
						Name:  "test",
						Title: "Test command",
						Args:  tt.args,
						RunFunc: func(context.Context, cli.Output, []string) error {
							return nil
						},
					},
				},
			}
			buf := &bytes.Buffer{}
			out := cli.NewWriter(buf)
			if _, err := app.Run(ctx, out, []string{"test", "-h"}); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			expectedUsage := strings.TrimSpace("Usage:\n  app test "+tt.expectedArgsString) + " [flags]\n"
			if helpText := buf.String(); !strings.Contains(helpText, expectedUsage) {
				t.Fatalf("expected help text to contain %q, got %q", expectedUsage, helpText)
			}
		})
	}
}
