package application

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"testing"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestConfirmLoginDefaults(t *testing.T) {
	for _, tt := range []struct {
		name        string
		defaults    [][]string
		args        []string
		environment string
		want        bool
	}{
		{name: "defaults to true", want: true},
		{name: "persisted true", defaults: [][]string{{"set", "confirm-login", "true"}}, want: true},
		{name: "persisted false", defaults: [][]string{{"set", "confirm-login", "false"}}},
		{name: "disable after enabling", defaults: [][]string{{"set", "confirm-login", "true"}, {"set", "confirm-login", "false"}}},
		{name: "unset restores true", defaults: [][]string{{"set", "confirm-login", "false"}, {"unset", "confirm-login"}}, want: true},
		{name: "flag overrides true", defaults: [][]string{{"set", "confirm-login", "true"}}, args: []string{"--confirm-login=false"}},
		{name: "flag overrides false", defaults: [][]string{{"set", "confirm-login", "false"}}, args: []string{"--confirm-login"}, want: true},
		{name: "environment disables confirmation", environment: "false"},
		{name: "environment enables confirmation", defaults: [][]string{{"set", "confirm-login", "false"}}, environment: "true", want: true},
		{name: "flag overrides environment", environment: "false", args: []string{"--confirm-login"}, want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NAIS_CONFIRM_LOGIN", tt.environment)
			config := filepath.Join(t.TempDir(), "config.yaml")
			for _, command := range tt.defaults {
				app, _, err := New(io.Discard)
				if err != nil {
					t.Fatal(err)
				}
				args := append([]string{"--config", config, "defaults"}, command...)
				if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
					t.Fatal(err)
				}
			}

			// Check the effective value inside a hook, after configuration and flags are synchronized.
			stop := errors.New("stop before command execution")
			called := false
			app, _, err := newApplication(io.Discard, func(global *flags.GlobalFlags) naistrix.PreRunHook {
				return naistrix.PreRunHook{
					Name: auth.NaisLoginHookName,
					RunFunc: func(context.Context, *naistrix.OutputWriter) error {
						called = true
						if global.ConfirmLogin != tt.want {
							t.Errorf("confirm-login = %v, want %v", global.ConfirmLogin, tt.want)
						}
						return stop
					},
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", config, "status"}, tt.args...)
			if err := app.Run(naistrix.RunWithArgs(args)); !errors.Is(err, stop) {
				t.Fatalf("error = %v, want hook sentinel", err)
			}
			if !called {
				t.Error("authentication hook did not run")
			}
		})
	}
}

func TestCommandLoginHooks(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want []string
	}{
		{name: "Nais API", args: []string{"status"}, want: []string{auth.NaisLoginHookName}},
		{name: "Nais Postgres", args: []string{"postgres", "list"}, want: []string{auth.NaisLoginHookName}},
		{name: "application alias", args: []string{"apps", "list"}, want: []string{auth.NaisLoginHookName}},
		{name: "workload metadata", args: []string{"auth", "workload-identity-metadata"}, want: []string{auth.NaisLoginHookName}},
		{name: "Cloud SQL listing uses API", args: []string{"cloudsql", "list"}, want: []string{auth.NaisLoginHookName}},
		{name: "Cloud SQL grant uses Google", args: []string{"cloudsql", "grant"}, want: []string{auth.GcloudLoginHookName}},
		{name: "Cloud SQL secrets and connection", args: []string{"cloudsql", "users", "list"}, want: []string{auth.NaisLoginHookName, auth.GcloudLoginHookName}},
		{name: "kubeconfig", args: []string{"kubeconfig"}, want: []string{auth.GcloudLoginHookName}},
		{name: "login alias", args: []string{"login"}},
		{name: "Nais login alias", args: []string{"login", "-n"}},
		{name: "nested login", args: []string{"auth", "login"}},
		{name: "logout alias", args: []string{"logout"}},
		{name: "token", args: []string{"auth", "token"}},
		{name: "auth status", args: []string{"auth", "status"}},
		{name: "auth status without login confirmation", args: []string{"auth", "status", "--confirm-login=false"}},
		{name: "auth status JSON", args: []string{"auth", "status", "--output", "json"}},
		{name: "auth status output completion", args: []string{"__complete", "auth", "status", "--output", ""}},
		{name: "validate", args: []string{"validate"}},
		{name: "device", args: []string{"device", "status"}},
		{name: "debug", args: []string{"debug"}},
		{name: "krakend", args: []string{"alpha", "krakend", "convert"}},
		{name: "MCP stdio", args: []string{"alpha", "mcp", "serve"}},
		{name: "defaults", args: []string{"defaults", "list"}},
		{name: "help", args: []string{"status", "--help"}},
		{name: "version", args: []string{"--version"}},
		{name: "subcommands", args: []string{"postgres"}},
		{name: "completion", args: []string{"__complete", "postgres", ""}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var called []string
			app, _, err := newApplication(io.Discard, func(*flags.GlobalFlags) naistrix.PreRunHook {
				return naistrix.PreRunHook{Name: auth.NaisLoginHookName, RunFunc: func(context.Context, *naistrix.OutputWriter) error {
					called = append(called, auth.NaisLoginHookName)
					return nil
				}}
			})
			if err != nil {
				t.Fatal(err)
			}
			// Exercise real command metadata and aliases without performing command side effects.
			var stub func([]*naistrix.Command)
			stub = func(commands []*naistrix.Command) {
				for _, cmd := range commands {
					for i, hook := range cmd.PreRunHooks {
						cmd.PreRunHooks[i].RunFunc = func(context.Context, *naistrix.OutputWriter) error {
							called = append(called, hook.Name)
							return nil
						}
					}
					cmd.Args = nil
					cmd.ValidateArgsFunc = nil
					cmd.ValidateFlagsFunc = nil
					if cmd.RunFunc != nil {
						cmd.RunFunc = func(context.Context, *naistrix.Arguments, *naistrix.OutputWriter) error { return nil }
					}
					stub(cmd.SubCommands)
				}
			}
			stub(app.Commands)
			args := append([]string{"--config", filepath.Join(t.TempDir(), "config.yaml")}, tt.args...)
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(called, tt.want) {
				t.Errorf("hooks = %v, want %v", called, tt.want)
			}
		})
	}
}
