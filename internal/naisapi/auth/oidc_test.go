package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"golang.org/x/oauth2"
)

type testOIDCProvider struct {
	issuer string
	domain string
	key    *rsa.PrivateKey
}

// startTestOIDCProvider's token endpoint issues the access token "new-access-token" with a fresh ID token.
func startTestOIDCProvider(t *testing.T, domain string) *testOIDCProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := jwk.Import(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	keys := jwk.NewSet()
	if err := errors.Join(
		publicKey.Set(jwk.KeyIDKey, "test"),
		publicKey.Set(jwk.AlgorithmKey, jwa.RS256()),
		keys.AddKey(publicKey),
	); err != nil {
		t.Fatal(err)
	}

	provider := &testOIDCProvider{domain: domain, key: key}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer": provider.issuer, "jwks_uri": provider.issuer + "/keys",
				"authorization_endpoint": provider.issuer + "/authorize", "token_endpoint": provider.issuer + "/token",
			})
		case "/keys":
			_ = json.NewEncoder(w).Encode(keys)
		case "/token":
			idToken, err := provider.signIDToken(time.Hour)
			if err != nil {
				t.Errorf("sign ID token: %v", err)
				http.Error(w, "sign ID token", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "new-access-token", "token_type": "Bearer", "expires_in": 3600,
				"refresh_token": "refresh-token", "id_token": idToken,
			})
		default:
			t.Errorf("unexpected OIDC request: %s", r.URL)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	provider.issuer = server.URL
	t.Setenv("NAIS_ZITADEL_DOMAIN", provider.issuer)
	t.Setenv("NAIS_ZITADEL_CLIENT_ID", "test-client")
	return provider
}

// token returns the access token "access-token", valid for an hour, and an ID token valid for idTokenLifetime.
func (p *testOIDCProvider) token(t *testing.T, idTokenLifetime time.Duration) *oauth2.Token {
	t.Helper()
	idToken, err := p.signIDToken(idTokenLifetime)
	if err != nil {
		t.Fatal(err)
	}
	return (&oauth2.Token{
		AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour),
	}).WithExtra(map[string]any{"id_token": idToken})
}

func (p *testOIDCProvider) signIDToken(lifetime time.Duration) (string, error) {
	idToken := jwt.New()
	for name, value := range map[string]any{
		"iss": p.issuer, "aud": "test-client", "exp": time.Now().Add(lifetime), "sub": "user-subject",
		"email": "user@" + p.domain, "email_verified": true,
		"urn:zitadel:iam:user:resourceowner:primary_domain": p.domain,
	} {
		if err := idToken.Set(name, value); err != nil {
			return "", err
		}
	}
	headers := jws.NewHeaders()
	if err := headers.Set(jws.KeyIDKey, "test"); err != nil {
		return "", err
	}
	signed, err := jwt.Sign(idToken, jwt.WithKey(jwa.RS256(), p.key, jws.WithProtectedHeaders(headers)))
	return string(signed), err
}

// interceptTokenEndpoint replaces http.DefaultClient, so callers can't use t.Parallel.
func interceptTokenEndpoint(t *testing.T, tokenFn func(*http.Request) (*http.Response, error)) {
	t.Helper()
	oldClient := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = oldClient })
	http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/token" {
			return tokenFn(r)
		}
		return http.DefaultTransport.RoundTrip(r)
	})}
}

func countTokenRequests(t *testing.T) *atomic.Int64 {
	t.Helper()
	var count atomic.Int64
	interceptTokenEndpoint(t, func(r *http.Request) (*http.Response, error) {
		count.Add(1)
		return http.DefaultTransport.RoundTrip(r)
	})
	return &count
}

func jsonResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}

func TestOIDCTenantSelection(t *testing.T) {
	for _, tc := range []struct {
		name      string
		domain    string
		tenant    string
		wantHost  string
		refresh   bool
		wantError bool
	}{
		{name: "tenant user unchanged", domain: "nav.no", wantHost: "console.nav.cloud.nais.io"},
		{name: "tenant user ignores override", domain: "nav.no", tenant: "dev-nais", wantHost: "console.nav.cloud.nais.io"},
		{name: "nais.io user selects nav", domain: "nais.io", tenant: "nav", wantHost: "console.nav.cloud.nais.io"},
		{name: "nais.io user selects dev-nais", domain: "nais.io", tenant: "dev-nais", wantHost: "console.dev-nais.cloud.nais.io"},
		{name: "nais.io session refresh preserves selection", domain: "nais.io", tenant: "dev-nais", wantHost: "console.dev-nais.cloud.nais.io", refresh: true},
		{name: "nais.io user needs tenant configuration not reauthentication", domain: "nais.io", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NAIS_API_TENANT", tc.tenant)
			setupCredentialStore(t)
			tok := startTestOIDCProvider(t, tc.domain).token(t, time.Hour)
			if tc.refresh {
				tok.Expiry = time.Now().Add(-time.Hour)
			}
			if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
				t.Fatal(err)
			}
			user, err := OIDC(t.Context())
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "NAIS_API_TENANT") || errors.Is(err, ErrNeedsOIDCLogin) {
					t.Fatalf("expected tenant configuration error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if user.ConsoleHost() != tc.wantHost || user.APIURL() != "https://"+tc.wantHost+"/graphql" || user.Domain() != tc.domain || user.Email() != "user@"+tc.domain {
				t.Fatalf("unexpected authenticated user: %+v", user)
			}
			if tc.refresh {
				if accessToken, err := user.AccessToken(); err != nil || accessToken != "new-access-token" {
					t.Fatalf("expected refreshed access token, got error %v", err)
				}
			}
		})
	}
}

func TestOIDCSwitchTenantWithoutLogin(t *testing.T) {
	setupCredentialStore(t)
	tok := startTestOIDCProvider(t, "nais.io").token(t, time.Hour)
	if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []string{"nav", "dev-nais", "nav"} {
		t.Setenv("NAIS_API_TENANT", tenant)
		user, err := OIDC(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if want := "console." + tenant + ".cloud.nais.io"; user.ConsoleHost() != want {
			t.Fatalf("ConsoleHost() = %q, want %q", user.ConsoleHost(), want)
		}
	}
}

func TestOIDCCachesAccessTokenNearExpiry(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lifetime time.Duration
	}{
		{name: "token between default and former expiry buffers", lifetime: 20 * time.Second},
		{name: "token well before expiry", lifetime: 2 * time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupCredentialStore(t)
			tok := startTestOIDCProvider(t, "nav.no").token(t, time.Hour)
			tok.Expiry = time.Now().Add(tc.lifetime)
			if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
				t.Fatal(err)
			}

			user, err := OIDC(t.Context())
			if err != nil {
				t.Fatal(err)
			}

			// Any reload from disk now fails, so the token must come from the cache.
			path, err := getCredentialsFilePath()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}

			for i := range 2 {
				accessToken, err := user.AccessToken()
				if err != nil || accessToken != "access-token" {
					t.Fatalf("AccessToken() call %d = %q, %v; want cached %q", i+1, accessToken, err, "access-token")
				}
			}
		})
	}
}

func TestOIDCCallbackTenantSelection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		domain string
		tenant string
		status int
		host   string
	}{
		{name: "nais.io login", domain: "nais.io", tenant: "nav", status: http.StatusOK, host: "console.nav.cloud.nais.io"},
		{name: "nais.io missing tenant", domain: "nais.io", status: http.StatusBadRequest},
		{name: "nais.io invalid tenant", domain: "nais.io", tenant: "../nav", status: http.StatusBadRequest},
		{name: "ordinary user ignores override", domain: "nav.no", tenant: "dev-nais", status: http.StatusOK, host: "console.nav.cloud.nais.io"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NAIS_API_TENANT", tc.tenant)
			setupCredentialStore(t)
			startTestOIDCProvider(t, tc.domain)
			oldClient := http.DefaultClient
			t.Cleanup(func() { http.DefaultClient = oldClient })
			http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host == "storage.googleapis.com" {
					if tc.domain == "nais.io" || r.URL.Path != "/nais-tenant-data/nav.no.json" {
						t.Fatalf("unexpected tenant metadata request: %s", r.URL)
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"consoleUrl":"console.nav.cloud.nais.io"}`)),
					}, nil
				}
				return http.DefaultTransport.RoundTrip(r)
			})}
			client, err := newOidcClient(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancelCause(t.Context())
			defer cancel(nil)
			srv := client.CallbackServer(ctx, cancel, "verifier", "state")
			response := httptest.NewRecorder()
			srv.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/callback?state=state&code=code", nil))
			if response.Code != tc.status {
				t.Fatalf("callback status=%d, want=%d: %s", response.Code, tc.status, response.Body.String())
			}
			if tc.status != http.StatusOK {
				if cause := context.Cause(ctx); cause == nil || !strings.Contains(cause.Error(), "NAIS_API_TENANT") {
					t.Fatalf("callback error was not propagated to CLI login: %v", cause)
				}
				if !strings.Contains(response.Body.String(), "NAIS_API_TENANT") {
					t.Fatalf("missing tenant configuration guidance: %s", response.Body.String())
				}
				path, _ := getCredentialsFilePath()
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("failed login unexpectedly stored credentials: %v", err)
				}
				return
			}
			stored, err := getOIDCUser(t.Context())
			if err != nil || stored.ConsoleHost != tc.host {
				t.Fatalf("unexpected stored tenant: %v, %v", stored, err)
			}
		})
	}
}

func TestOIDCJWKSOutageIsNotNeedsLogin(t *testing.T) {
	setupCredentialStore(t)
	tok := startTestOIDCProvider(t, "nav.no").token(t, time.Hour)
	if _, err := storeOIDCUser(t.Context(), tok, "console.nav.cloud.nais.io"); err != nil {
		t.Fatal(err)
	}

	// jwx fetches JWKS through its package-level client; restore the default on cleanup.
	t.Cleanup(func() { jwk.Configure(jwk.WithHTTPClient(jwk.DefaultHTTPClient())) })
	jwk.Configure(jwk.WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/keys" {
			return nil, errors.New("simulated JWKS outage")
		}
		return http.DefaultTransport.RoundTrip(r)
	})}))

	_, err := OIDC(t.Context())

	if err == nil || errors.Is(err, ErrNeedsOIDCLogin) {
		t.Fatalf("OIDC() error = %v, want an error that does not require login", err)
	}
}
