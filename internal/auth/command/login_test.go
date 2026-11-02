package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func TestOfferNaisLogin(t *testing.T) {
	checkError := errors.New("authentication service unavailable")
	promptError := errors.New("prompt cancelled")
	loginError := errors.New("login failed")
	for _, tt := range []struct {
		name                                      string
		checkErr, promptErr, loginErr             error
		autoAccept, accept, wantPrompt, wantLogin bool
		wantError                                 error
	}{
		{name: "already authenticated"},
		{name: "already authenticated with auto-accept", autoAccept: true},
		{name: "login required and accepted", checkErr: fmt.Errorf("expired: %w", naisapi.ErrNeedsLogin), accept: true, wantPrompt: true, wantLogin: true},
		{name: "login required and declined", checkErr: naisapi.ErrNeedsLogin, wantPrompt: true},
		{name: "auto-accept skips prompt when login required", checkErr: naisapi.ErrNeedsLogin, autoAccept: true, wantLogin: true},
		{name: "unrelated check failure", checkErr: checkError, wantError: checkError},
		{name: "unrelated check failure with auto-accept", checkErr: checkError, autoAccept: true, wantError: checkError},
		{name: "prompt failure", checkErr: naisapi.ErrNeedsLogin, promptErr: promptError, wantPrompt: true, wantError: promptError},
		{name: "login failure", checkErr: naisapi.ErrNeedsLogin, accept: true, loginErr: loginError, wantPrompt: true, wantLogin: true, wantError: loginError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			level := naistrix.OutputVerbosityLevelNormal
			out := naistrix.NewOutputWriter(&output, &level)
			ctx := t.Context()
			checks, prompts, logins := 0, 0, 0
			err := offerNaisLogin(ctx, out, tt.autoAccept, func(got context.Context) (naisapi.AuthenticatedUser, error) {
				checks++
				if got != ctx {
					t.Error("authentication check received the wrong context")
				}
				return nil, tt.checkErr
			}, func(got context.Context, gotOut *naistrix.OutputWriter) error {
				logins++
				if checks != 1 || got != ctx || gotOut != out {
					t.Error("login ran before the session check or received incorrect arguments")
				}
				if !tt.autoAccept && prompts != 1 {
					t.Error("login ran before confirmation")
				}
				return tt.loginErr
			}, func(prompt string, opts ...input.ConfirmOptionFunc) (bool, error) {
				prompts++
				if checks != 1 {
					t.Error("confirmation ran before the session check")
				}
				if prompt != "Would you like to also log in to Nais?" || len(opts) != 1 {
					t.Errorf("unexpected confirmation prompt or options: %q, %d options", prompt, len(opts))
				}
				return tt.accept, tt.promptErr
			})
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("error = %v, want %v", err, tt.wantError)
			}
			wantPrompts, wantLogins := 0, 0
			if tt.wantPrompt {
				wantPrompts = 1
			}
			if tt.wantLogin {
				wantLogins = 1
			}
			if checks != 1 || prompts != wantPrompts || logins != wantLogins {
				t.Errorf("checks/prompts/logins = %d/%d/%d, want 1/%d/%d", checks, prompts, logins, wantPrompts, wantLogins)
			}
			if !errors.Is(tt.checkErr, naisapi.ErrNeedsLogin) && output.Len() != 0 {
				t.Errorf("unexpected Nais login message: %q", output.String())
			}
		})
	}
}
