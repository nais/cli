package cli_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nais/cli/pkg/cli"
)

var discard = cli.NewWriter(io.Discard)

func TestCommandValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		cmd           *cli.Command
		panicContains string
	}{
		{
			name: "command with no name",
			cmd: &cli.Command{
				Title: "Test command",
			},
			panicContains: "cannot be empty",
		},
		{
			name: "command with space in name",
			cmd: &cli.Command{
				Name:  "test command",
				Title: "Test command",
			},
			panicContains: "cannot contain spaces: test command",
		},
		{
			name: "command with no title",
			cmd: &cli.Command{
				Name: "cmd",
			},
			panicContains: "missing a title",
		},
		{
			name: "command with newline in title",
			cmd: &cli.Command{
				Name:  "test",
				Title: "Test command\nwith newline",
			},
			panicContains: "contains newline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Application{
				Name:        "app",
				SubCommands: []*cli.Command{tt.cmd},
			}

			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for command with no name, but did not panic")
				} else if msg := r.(string); !strings.Contains(msg, tt.panicContains) {
					t.Fatalf("expected panic message to contain %q, got: %q", tt.panicContains, msg)
				}
			}()

			_, _ = app.Run(ctx, discard, []string{"-h"})
		})
	}
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
			name:               "argument",
			expectedArgsString: "ARG",
			args: []cli.Argument{
				{Name: "arg"},
			},
		},
		{
			name:               "repeatable argument",
			expectedArgsString: "ARG [ARG...]",
			args: []cli.Argument{
				{Name: "arg", Repeatable: true},
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

func TestCommandArgumentValidation(t *testing.T) {
	ctx := context.Background()
	out := cli.NewWriter(io.Discard)

	tests := []struct {
		name          string
		args          []cli.Argument
		panicContains string
	}{
		{
			name: "missing argument name",
			args: []cli.Argument{
				{Repeatable: true},
			},
			panicContains: "cannot be empty",
		},
		{
			name: "repeatable argument must be last",
			args: []cli.Argument{
				{Name: "arg1", Repeatable: true},
				{Name: "arg2"},
			},
			panicContains: "must be the last argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if msg, ok := recover().(string); ok && !strings.Contains(msg, tt.panicContains) {
					t.Fatalf("expected panic message to contain %q, got: %q", tt.panicContains, msg)
				}
			}()
			_, _ = (&cli.Application{
				Name: "app",
				SubCommands: []*cli.Command{
					{
						Name:  "test",
						Title: "Test command",
						Args:  tt.args,
					},
				},
			}).Run(ctx, out, []string{"-h"})
			t.Fatalf("expected panic")
		})
	}
}
