package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
	naisapiauth "github.com/nais/cli/internal/naisapi/auth"
	"github.com/nais/naistrix"
)

func TestLoginHook(t *testing.T) {
	otherError := errors.New("authentication service unavailable")
	loginError := errors.New("login cancelled")
	promptError := errors.New("prompt cancelled")
	for _, tt := range []struct {
		name         string
		checkError   error
		loginError   error
		promptError  error
		decline      bool
		interactive  bool
		confirmLogin bool
		wantLogin    bool
		wantError    error
	}{
		{name: "already authenticated", interactive: true, confirmLogin: true},
		{name: "login before execution", checkError: fmt.Errorf("expired: %w", naisapi.ErrNeedsLogin), interactive: true, confirmLogin: true, wantLogin: true},
		{name: "noninteractive", checkError: naisapi.ErrNeedsLogin, confirmLogin: true, wantError: naisapi.ErrNeedsLogin},
		{name: "unrelated error", checkError: otherError, interactive: true, confirmLogin: true, wantError: otherError},
		{name: "failed login", checkError: naisapi.ErrNeedsLogin, interactive: true, confirmLogin: true, loginError: loginError, wantLogin: true, wantError: loginError},
		{name: "declined login", checkError: naisapi.ErrNeedsLogin, interactive: true, confirmLogin: true, decline: true, wantError: naisapi.ErrNeedsLogin},
		{name: "failed prompt", checkError: naisapi.ErrNeedsLogin, interactive: true, confirmLogin: true, promptError: promptError, wantError: promptError},
		{name: "confirmation disabled", checkError: naisapi.ErrNeedsLogin, interactive: true, wantLogin: true},
		{name: "confirmation disabled stays noninteractive", checkError: naisapi.ErrNeedsLogin, wantError: naisapi.ErrNeedsLogin},
		{name: "confirmation disabled already authenticated", interactive: true},
		{name: "confirmation disabled unrelated error", checkError: otherError, interactive: true, wantError: otherError},
		{name: "confirmation disabled failed login", checkError: naisapi.ErrNeedsLogin, interactive: true, loginError: loginError, wantLogin: true, wantError: loginError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			checks, logins, choices, runs := 0, 0, 0, 0
			prompts := 0
			ctx := t.Context()
			hook := loginHook("auth", "Nais", naisapi.ErrNeedsLogin, func(got context.Context) error {
				checks++
				if got != ctx {
					t.Error("check did not receive execution context")
				}
				return tt.checkError
			}, func(got context.Context, _ *naistrix.OutputWriter) error {
				if tt.confirmLogin && prompts != 1 {
					t.Error("login started before confirmation")
				}
				logins++
				if got != ctx {
					t.Error("login did not receive execution context")
				}
				return tt.loginError
			}, func() bool { return false }, func() bool { return tt.interactive }, func() bool { return tt.confirmLogin }, func(prompt string) (bool, error) {
				if prompt != "Log in to Nais to continue?" {
					t.Errorf("unexpected prompt: %q", prompt)
				}
				prompts++
				return !tt.decline, tt.promptError
			})
			app, _, err := naistrix.NewApplication("test", "Test", "test", naistrix.ApplicationWithWriter(&output), naistrix.ApplicationWithPreRunHooks(hook))
			if err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(&naistrix.Command{
				Name: "run", Title: "Run",
				Args: []naistrix.Argument{{Name: "name", Prompt: "Name", ChoicesFunc: func(context.Context, *naistrix.Arguments) ([]string, error) {
					choices++
					if tt.wantLogin && logins != 1 {
						t.Error("argument choices ran before login")
					}
					return []string{"example"}, nil
				}}},
				RunFunc: func(context.Context, *naistrix.Arguments, *naistrix.OutputWriter) error {
					runs++
					return nil
				},
			}); err != nil {
				t.Fatal(err)
			}
			err = app.Run(naistrix.RunWithContext(ctx), naistrix.RunWithArgs([]string{"--config", t.TempDir() + "/config.yaml", "run", "example"}))
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("error = %v, want %v", err, tt.wantError)
			}
			wantLogins := 0
			wantPrompts := 0
			if tt.confirmLogin && tt.interactive && errors.Is(tt.checkError, naisapi.ErrNeedsLogin) {
				wantPrompts = 1
			}
			if prompts != wantPrompts {
				t.Errorf("prompts = %d, want %d", prompts, wantPrompts)
			}
			if hint := strings.Contains(output.String(), "nais defaults set confirm-login false"); hint != (wantPrompts == 1) {
				t.Errorf("confirmation hint visible = %v, want %v", hint, wantPrompts == 1)
			}
			if tt.wantLogin {
				wantLogins = 1
			}
			wantRuns := 0
			if tt.wantError == nil {
				wantRuns = 1
			}
			if checks != 1 || logins != wantLogins || runs != wantRuns || choices != wantRuns {
				t.Errorf("checks/logins/choices/runs = %d/%d/%d/%d, want 1/%d/%d/%d", checks, logins, choices, runs, wantLogins, wantRuns, wantRuns)
			}
			if wantPrompts == 0 && strings.Contains(output.String(), "authentication required") {
				t.Errorf("unexpected authentication announcement: %s", output.String())
			}
			if !tt.confirmLogin && output.Len() != 0 && tt.wantError == nil {
				t.Errorf("unexpected output with confirmation disabled: %s", output.String())
			}
		})
	}
}

func TestLoginHookVerboseReason(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		t.Run(fmt.Sprint(verbose), func(t *testing.T) {
			var output bytes.Buffer
			level := naistrix.OutputVerbosityLevelNormal
			hook := loginHook("auth", "Nais", naisapi.ErrNeedsLogin, func(context.Context) error {
				return &naisapiauth.NeedsLoginError{Cause: errors.New("refresh token rejected")}
			}, func(context.Context, *naistrix.OutputWriter) error { return nil }, func() bool { return verbose }, func() bool { return true }, func() bool { return true }, func(string) (bool, error) { return true, nil })
			if err := hook.RunFunc(t.Context(), naistrix.NewOutputWriter(&output, &level)); err != nil {
				t.Fatal(err)
			}
			if got := strings.Contains(output.String(), "refresh token rejected"); got != verbose {
				t.Errorf("reason visible = %v, want %v: %s", got, verbose, output.String())
			}
		})
	}
}

func TestLoginHookProviderMessages(t *testing.T) {
	for _, tt := range []struct {
		name, provider, prompt string
		needsLogin             error
	}{
		{
			name: NaisLoginHookName, provider: "Nais", needsLogin: naisapi.ErrNeedsLogin,
			prompt: "Log in to Nais to continue?",
		},
		{
			name: GcloudLoginHookName, provider: "Google Cloud", needsLogin: gcloud.ErrNeedsLogin,
			prompt: "Log in to Google Cloud to continue?",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			level := naistrix.OutputVerbosityLevelNormal
			var prompt string
			hook := loginHook(tt.name, tt.provider, tt.needsLogin, func(context.Context) error {
				return tt.needsLogin
			}, func(context.Context, *naistrix.OutputWriter) error { return nil }, func() bool { return false }, func() bool { return true }, func() bool { return true }, func(text string) (bool, error) {
				prompt = text
				return true, nil
			})
			if err := hook.RunFunc(t.Context(), naistrix.NewOutputWriter(&output, &level)); err != nil {
				t.Fatal(err)
			}
			if prompt != tt.prompt {
				t.Errorf("prompt = %q, want %q", prompt, tt.prompt)
			}
			if output.String() != "Skip prompts: `nais defaults set confirm-login false`.\n" {
				t.Errorf("output = %q, want only the confirmation hint", output.String())
			}
		})
	}
}
