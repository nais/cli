package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestResolveConsoleHost(t *testing.T) {
	for _, tc := range []struct {
		name       string
		domain     string
		storedHost string
		tenant     string
		want       string
		wantError  bool
	}{
		{name: "normal user keeps automatic tenant", domain: "nav.no", storedHost: "console.nav.cloud.nais.io", want: "console.nav.cloud.nais.io"},
		{name: "normal user ignores override", domain: "nav.no", storedHost: "console.nav.cloud.nais.io", tenant: "dev-nais", want: "console.nav.cloud.nais.io"},
		{name: "normal user ignores invalid override", domain: "nav.no", storedHost: "console.nav.cloud.nais.io", tenant: "https://invalid.example", want: "console.nav.cloud.nais.io"},
		{name: "nais.io selects nav", domain: "nais.io", tenant: "nav", want: "console.nav.cloud.nais.io"},
		{name: "nais.io selects dev-nais", domain: "nais.io", tenant: "dev-nais", want: "console.dev-nais.cloud.nais.io"},
		{name: "nais.io override replaces stored tenant", domain: "nais.io", storedHost: "console.nav.cloud.nais.io", tenant: "dev-nais", want: "console.dev-nais.cloud.nais.io"},
		{name: "nais.io requires explicit tenant", domain: "nais.io", storedHost: "console.nav.cloud.nais.io", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NAIS_API_TENANT", tc.tenant)
			got, err := resolveConsoleHost(t.Context(), tc.domain, tc.storedHost)
			if (err != nil) != tc.wantError || got != tc.want {
				t.Fatalf("resolveConsoleHost() = %q, %v; want %q, error=%v", got, err, tc.want, tc.wantError)
			}
			if tc.wantError && !strings.Contains(err.Error(), "NAIS_API_TENANT") {
				t.Fatalf("missing tenant configuration guidance: %v", err)
			}
		})
	}
}

func TestResolveConsoleHostRejectsInvalidTenantNames(t *testing.T) {
	for _, tenant := range []string{
		"nav.no", "NAV", "../nav", "nav/path", "nav:443", "user@nav", "nav?query", "nav#fragment",
		"-nav", "nav-", " nav", "nav ", strings.Repeat("n", 64), "https://console.nav.cloud.nais.io",
	} {
		t.Run(tenant, func(t *testing.T) {
			t.Setenv("NAIS_API_TENANT", tenant)
			if _, err := resolveConsoleHost(t.Context(), "nais.io", ""); err == nil {
				t.Fatal("expected invalid tenant name to be rejected")
			}
		})
	}
}

func TestResolveConsoleHostDiscoversNormalTenant(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "tenant metadata", status: http.StatusOK, body: `{"consoleUrl":"console.nav.cloud.nais.io"}`, want: "console.nav.cloud.nais.io"},
		{name: "tenant not found", status: http.StatusNotFound, body: `{}`},
		{name: "metadata unavailable", status: http.StatusServiceUnavailable, body: `{}`},
		{name: "invalid metadata", status: http.StatusOK, body: `invalid json`},
		{name: "missing console host", status: http.StatusOK, body: `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NAIS_API_TENANT", "dev-nais")
			oldClient := http.DefaultClient
			t.Cleanup(func() { http.DefaultClient = oldClient })
			http.DefaultClient = &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://storage.googleapis.com/nais-tenant-data/nav.no.json" {
					t.Fatalf("unexpected tenant metadata request: %s", r.URL)
				}
				return &http.Response{
					StatusCode: tc.status,
					Status:     fmt.Sprintf("%d %s", tc.status, http.StatusText(tc.status)),
					Body:       io.NopCloser(strings.NewReader(tc.body)),
				}, nil
			})}
			got, err := resolveConsoleHost(context.Background(), "nav.no", "")
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("resolveConsoleHost() = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
