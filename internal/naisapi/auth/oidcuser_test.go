package auth

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestConcurrentCallersRefreshOnce(t *testing.T) {
	setupCredentialStore(t)
	tok := startTestOIDCProvider(t, "nav.no").token(t, time.Hour)
	tok.Expiry = time.Now().Add(-time.Hour)
	if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
		t.Fatal(err)
	}
	requests := countTokenRequests(t)

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range cap(errs) {
		wg.Go(func() {
			user, err := getOIDCUser(t.Context())
			if err == nil && user.AccessToken != "new-access-token" {
				err = fmt.Errorf("access token = %q, want refreshed token", user.AccessToken)
			}
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("token endpoint requests = %d, want 1", got)
	}
}

func TestOIDCRefreshFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		name      string
		tokenFn   func(*http.Request) (*http.Response, error)
		wantLogin bool
	}{
		{
			name: "transport error",
			tokenFn: func(*http.Request) (*http.Response, error) {
				return nil, errors.New("simulated token endpoint transport error")
			},
		},
		{
			name: "server error",
			tokenFn: func(r *http.Request) (*http.Response, error) {
				return jsonResponse(r, http.StatusServiceUnavailable, `{"error":"temporarily_unavailable"}`), nil
			},
		},
		{
			name: "rejected refresh token",
			tokenFn: func(r *http.Request) (*http.Response, error) {
				return jsonResponse(r, http.StatusBadRequest, `{"error":"invalid_grant"}`), nil
			},
			wantLogin: true,
		},
		{
			name: "zitadel rejected refresh token",
			tokenFn: func(r *http.Request) (*http.Response, error) {
				return jsonResponse(r, http.StatusBadRequest, `{"error":"invalid_request","error_description":"Errors.OIDCSession.RefreshTokenInvalid"}`), nil
			},
			wantLogin: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupCredentialStore(t)
			tok := startTestOIDCProvider(t, "nav.no").token(t, time.Hour)
			tok.Expiry = time.Now().Add(-time.Hour)
			if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
				t.Fatal(err)
			}
			interceptTokenEndpoint(t, tc.tokenFn)

			_, err := getOIDCUser(t.Context())

			if err == nil {
				t.Fatal("getOIDCUser() succeeded, want error")
			}
			if got := errors.Is(err, ErrNeedsOIDCLogin); got != tc.wantLogin {
				t.Fatalf("errors.Is(ErrNeedsOIDCLogin) = %v, want %v: %v", got, tc.wantLogin, err)
			}
			if tc.wantLogin {
				rejected, ok := errors.AsType[*NeedsLoginError](err)
				if !ok || rejected.Cause == nil || err.Error() != ErrNeedsOIDCLogin.Error() {
					t.Fatalf("error = %q, want only the login instruction, with the provider's error as its cause", err)
				}
			}
		})
	}
}
