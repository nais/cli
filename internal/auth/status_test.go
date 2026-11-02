package auth

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/cli/internal/naisapi"
)

func TestLoginStatuses(t *testing.T) {
	unavailable := errors.New("identity provider unavailable")
	gcloudFailed := errors.New("gcloud configuration could not be read")
	for _, tt := range []struct {
		name                     string
		naisErr, googleErr       error
		naisStatus, googleStatus string
		wantError                bool
	}{
		{name: "both ready", naisStatus: "Authenticated", googleStatus: "Authenticated"},
		{name: "both require login", naisErr: fmt.Errorf("expired: %w", naisapi.ErrNeedsLogin), googleErr: fmt.Errorf("missing ADC: %w", gcloud.ErrNeedsLogin), naisStatus: "Not logged in", googleStatus: "Not logged in"},
		{name: "Nais requires login", naisErr: naisapi.ErrNeedsLogin, naisStatus: "Not logged in", googleStatus: "Authenticated"},
		{name: "Google requires login", googleErr: gcloud.ErrNeedsLogin, naisStatus: "Authenticated", googleStatus: "Not logged in"},
		{name: "Nais failure still checks Google", naisErr: unavailable, naisStatus: "Check failed", googleStatus: "Authenticated", wantError: true},
		{name: "Google failure still reports Nais", googleErr: gcloudFailed, naisStatus: "Authenticated", googleStatus: "Check failed", wantError: true},
		{name: "both checks fail", naisErr: unavailable, googleErr: gcloudFailed, naisStatus: "Check failed", googleStatus: "Check failed", wantError: true},
		{name: "gcloud not installed", googleErr: fmt.Errorf("lookup gcloud: %w", exec.ErrNotFound), naisStatus: "Authenticated", googleStatus: "Not installed"},
		{name: "missing Nais dependency is not missing gcloud", naisErr: exec.ErrNotFound, naisStatus: "Check failed", googleStatus: "Authenticated", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			calls := 0
			check := func(account string, err error) func(context.Context) (string, error) {
				return func(got context.Context) (string, error) {
					calls++
					if got != ctx {
						t.Error("checker did not receive execution context")
					}
					return account, err
				}
			}
			statuses, err := checkLoginStatuses(ctx, []loginStatusCheck{
				{provider: "Nais", needsLogin: naisapi.ErrNeedsLogin, check: check("nais@example.com", tt.naisErr)},
				{provider: "Google Cloud", needsLogin: gcloud.ErrNeedsLogin, check: check("google@example.com", tt.googleErr)},
			})
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}
			if calls != 2 || len(statuses) != 2 {
				t.Fatalf("calls = %d, statuses = %v; want both providers checked", calls, statuses)
			}
			for i, want := range []struct {
				provider, status, account string
				err                       error
			}{
				{"Nais", tt.naisStatus, "nais@example.com", tt.naisErr},
				{"Google Cloud", tt.googleStatus, "google@example.com", tt.googleErr},
			} {
				got := statuses[i]
				if got.Provider != want.provider || got.Status.String() != want.status {
					t.Errorf("status = %+v, want %s / %s", got, want.provider, want.status)
				}
				if want.err == nil {
					if got.Account == nil || *got.Account != want.account || got.Details != "" {
						t.Errorf("ready status = %+v, want account %s and no error details", got, want.account)
					}
				} else if got.Account != nil || got.Details == "" {
					t.Errorf("failed status = %+v, want no account and diagnostic details", got)
				}
				if want.status == "Check failed" && !errors.Is(err, want.err) {
					t.Errorf("error = %v, want original error %v preserved", err, want.err)
				}
			}
		})
	}
}

func TestLoginStatusWithoutAccount(t *testing.T) {
	statuses, err := checkLoginStatuses(t.Context(), []loginStatusCheck{{
		provider: "Nais", needsLogin: naisapi.ErrNeedsLogin,
		check: func(context.Context) (string, error) { return "", nil },
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].Account != nil || statuses[0].Status != LoginStateAuthenticated {
		t.Errorf("statuses = %+v, want authenticated status without an account", statuses)
	}
}
