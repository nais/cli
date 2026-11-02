package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestAuthStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script to stub gcloud")
	}
	for _, tt := range []struct {
		name, script, want, wantJSON string
		adc, wantError               bool
	}{
		{name: "both authenticated", script: "printf '%s\\n' 'google@example.com'", adc: true, want: "Authenticated", wantJSON: "authenticated"},
		{name: "missing account", script: "exit 0", want: "Not logged in", wantJSON: "not_logged_in"},
		{name: "missing ADC", script: "printf '%s\\n' 'google@example.com'", want: "Not logged in", wantJSON: "not_logged_in"},
		{name: "gcloud missing", want: "Not installed", wantJSON: "not_installed"},
		{name: "gcloud check failed", script: "exit 1", want: "Check failed", wantJSON: "check_failed", wantError: true},
	} {
		for _, format := range []string{"", "table", "json"} {
			t.Run(tt.name+"/output="+format, func(t *testing.T) {
				home := t.TempDir()
				t.Setenv("HOME", home)
				t.Setenv("PATH", home)
				t.Setenv("NAIS_API_LOCAL_HOST", "localhost:4242")
				t.Setenv("NAIS_API_LOCAL_EMAIL", "nais@example.com")
				t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
				if err := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS"); err != nil {
					t.Fatal(err)
				}
				argsFile := filepath.Join(home, "gcloud-args")
				t.Setenv("NAIS_TEST_GCLOUD_ARGS", argsFile)
				if tt.script != "" {
					script := "#!/bin/sh\nprintf '%s' \"$*\" > \"$NAIS_TEST_GCLOUD_ARGS\"\n" + tt.script + "\n"
					if err := os.WriteFile(filepath.Join(home, "gcloud"), []byte(script), 0o700); err != nil {
						t.Fatal(err)
					}
				}
				if tt.adc {
					dir := filepath.Join(home, ".config", "gcloud")
					if err := os.MkdirAll(dir, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, "application_default_credentials.json"), []byte(`{"refresh_token":"do-not-print-this-token"}`), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				var output bytes.Buffer
				app, global, err := naistrix.NewApplication("nais", "Test auth status", "test",
					naistrix.ApplicationWithWriter(&output),
					naistrix.ApplicationWithPreRunHooks(naistrix.PreRunHook{Name: auth.NaisLoginHookName, RunFunc: func(context.Context, *naistrix.OutputWriter) error {
						t.Error("auth status must not run the Nais login hook")
						return errors.New("unexpected login hook")
					}}),
				)
				if err != nil {
					t.Fatal(err)
				}
				additional := &flags.AdditionalFlags{}
				if err := app.AddGlobalFlags(additional); err != nil {
					t.Fatal(err)
				}
				if err := app.AddCommand(Auth(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
					t.Fatal(err)
				}
				args := []string{"--config", filepath.Join(home, "config.yaml"), "--no-colors", "auth", "status", "--confirm-login=false"}
				if format != "" {
					args = append(args, "--output", format)
				}
				err = app.Run(naistrix.RunWithArgs(args))
				if (err != nil) != tt.wantError {
					t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
				}
				if format == "json" {
					var report map[string]map[string]any
					if err := json.Unmarshal(output.Bytes(), &report); err != nil {
						t.Fatalf("output is not a single JSON document: %v\n%s", err, output.String())
					}
					if len(report) != 2 || report["nais"] == nil || report["gcloud"] == nil {
						t.Fatalf("unexpected provider keys: %v", report)
					}
					if report["nais"]["status"] != "authenticated" || report["nais"]["account"] != "nais@example.com" {
						t.Errorf("unexpected Nais status: %v", report["nais"])
					}
					if _, ok := report["nais"]["details"]; ok {
						t.Error("authenticated Nais status should omit diagnostic details")
					}
					google := report["gcloud"]
					if google["status"] != tt.wantJSON {
						t.Errorf("Google status = %v, want %s", google["status"], tt.wantJSON)
					}
					account, ok := google["account"]
					if !ok || (tt.adc && account != "google@example.com") || (!tt.adc && account != nil) {
						t.Errorf("Google account = %v (present %v), want account or explicit null", account, ok)
					}
					if tt.adc {
						if _, ok := google["details"]; ok {
							t.Error("authenticated Google status should omit diagnostic details")
						}
					} else if details, ok := google["details"].(string); !ok || details == "" {
						t.Errorf("missing diagnostic details: %v", google)
					}
				} else {
					for _, want := range []string{"Nais", "Authenticated", "nais@example.com", "Google Cloud", tt.want} {
						if !strings.Contains(output.String(), want) {
							t.Errorf("output = %q, want %q", output.String(), want)
						}
					}
					if tt.adc && !strings.Contains(output.String(), "google@example.com") {
						t.Error("missing Google Cloud account")
					}
					if tt.adc && strings.Count(output.String(), "Authenticated") != 2 {
						t.Errorf("expected both providers to be authenticated: %s", output.String())
					}
				}
				if tt.script != "" {
					args, err := os.ReadFile(argsFile)
					if err != nil {
						t.Fatal(err)
					}
					if string(args) != "config list account --format value(core.account)" {
						t.Errorf("unexpected gcloud invocation: %q", args)
					}
				}
				if strings.Contains(output.String(), "Skip prompts") || strings.Contains(output.String(), "to continue?") {
					t.Error("auth status displayed a login prompt")
				}
				for _, secret := range []string{"do-not-print-this-token", "not-really-an-access-token"} {
					if strings.Contains(output.String(), secret) {
						t.Error("authentication status leaked a token")
					}
				}
			})
		}
	}
}

func TestAuthStatusOutputValidation(t *testing.T) {
	for _, format := range []flag.StatusOutput{"table", "json", "yaml", "unknown"} {
		t.Run(string(format), func(t *testing.T) {
			cmd := status(&flag.Auth{})
			f := cmd.Flags.(*flag.Status)
			f.Output = format
			err := cmd.ValidateFlagsFunc(t.Context())
			valid := format == "table" || format == "json"
			if (err == nil) != valid {
				t.Errorf("error = %v, want valid = %v", err, valid)
			}
		})
	}
}

func TestAuthStatusOutputCompletion(t *testing.T) {
	var format flag.StatusOutput
	choices, _ := format.AutoComplete(t.Context(), nil, "", nil)
	if len(choices) != 2 || choices[0] != "table" || choices[1] != "json" {
		t.Errorf("choices = %v, want table and json", choices)
	}
}

func TestRenderAuthStatusJSONNullAccounts(t *testing.T) {
	var output bytes.Buffer
	level := naistrix.OutputVerbosityLevelNormal
	err := renderLoginStatuses(naistrix.NewOutputWriter(&output, &level), auth.LoginStatusReport{
		Nais:   auth.LoginStatus{Provider: "Nais", Status: auth.LoginStateNotLoggedIn, Details: "Run `nais login -n`."},
		Gcloud: auth.LoginStatus{Provider: "Google Cloud", Status: auth.LoginStateAuthenticated},
	}, "json")
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]map[string]any
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report["nais"]["status"] != "not_logged_in" || report["gcloud"]["status"] != "authenticated" {
		t.Errorf("unexpected status values: %s", output.String())
	}
	for _, provider := range []string{"nais", "gcloud"} {
		if account, ok := report[provider]["account"]; !ok || account != nil {
			t.Errorf("%s account = %v (present %v), want explicit null", provider, account, ok)
		}
	}
}
