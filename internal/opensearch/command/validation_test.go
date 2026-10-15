package command

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestOpenSearchValidationPhases(t *testing.T) {
	for _, tt := range []struct {
		name        string
		args        []string
		wantError   string
		completion  bool
		wantChoices bool
		wantRun     bool
	}{
		{
			name:      "parent flags validated before argument choices",
			args:      []string{"create", "instance", "-e", "dev"},
			wantError: "missing required team",
		},
		{
			name:      "environment validated before argument choices",
			args:      []string{"create", "instance", "-t", "my-team"},
			wantError: "missing required environment",
		},
		{
			name:      "command flags validated before missing arguments",
			args:      []string{"create", "-t", "my-team", "-e", "dev", "--memory", "invalid"},
			wantError: "invalid memory",
		},
		{
			name:        "valid flags allow argument choices and execution",
			args:        []string{"create", "instance", "-t", "my-team", "-e", "dev"},
			wantChoices: true,
			wantRun:     true,
		},
		{
			name:       "parent flags validated during automatic completion",
			args:       []string{"create", "-e", "dev", ""},
			wantError:  "missing required team",
			completion: true,
		},
		{
			name:       "command flags validated during automatic completion",
			args:       []string{"create", "-t", "my-team", "-e", "dev", "--memory", "invalid", ""},
			wantError:  "invalid memory",
			completion: true,
		},
		{
			name:        "completion does not validate incomplete arguments",
			args:        []string{"create", "-t", "my-team", "-e", "dev", ""},
			completion:  true,
			wantChoices: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			cmd := OpenSearch(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})
			create := cmd.SubCommands[0]
			choicesCalled, runCalled := false, false
			create.Args[0].ChoicesFunc = func(context.Context, *naistrix.Arguments) ([]string, error) {
				choicesCalled = true
				return []string{"instance"}, nil
			}
			create.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
				runCalled = true
				if got := args.Get("name"); got != "instance" {
					t.Errorf("name = %q, want instance", got)
				}
				return nil
			}
			if err := app.AddCommand(cmd); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"opensearch"}, tt.args...)
			if tt.completion {
				args = append([]string{"__complete"}, args...)
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.completion {
				if err != nil {
					t.Fatal(err)
				}
				if tt.wantError != "" {
					if got := output.String(); !strings.Contains(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) {
						t.Errorf("completion = %q, want active help containing %q", got, tt.wantError)
					}
				} else if !strings.Contains(output.String(), "instance\n") {
					t.Errorf("completion = %q, want instance suggestion", output.String())
				}
			} else if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if choicesCalled != tt.wantChoices {
				t.Errorf("choices called = %v, want %v", choicesCalled, tt.wantChoices)
			}
			if runCalled != tt.wantRun {
				t.Errorf("run called = %v, want %v", runCalled, tt.wantRun)
			}
		})
	}
}
