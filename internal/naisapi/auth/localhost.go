package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

type LocalhostUser struct {
	AuthenticatedUser
}

func Localhost() (*LocalhostUser, bool) {
	host := os.Getenv("NAIS_API_LOCAL_HOST")
	if host == "" {
		return nil, false
	}

	return &LocalhostUser{
		AuthenticatedUser{
			consoleHost: host,
			domain:      "api.nais.localhost",
			email:       os.Getenv("NAIS_API_LOCAL_EMAIL"),
			ts: oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: "not-really-an-access-token",
			}),
		},
	}, true
}

// APIURL overrides the parent method to use HTTP instead of HTTPS for local development
func (l *LocalhostUser) APIURL() string {
	return fmt.Sprintf("http://%s/graphql", l.ConsoleHost())
}

func (l *LocalhostUser) HTTPClient(_ context.Context) *http.Client {
	return &http.Client{
		Transport: l.RoundTripper(http.DefaultTransport),
	}
}

func (l *LocalhostUser) RoundTripper(base http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		_ = l.SetAuthorizationHeader(r.Header)
		return base.RoundTrip(r)
	})
}

func (l *LocalhostUser) SetAuthorizationHeader(headers http.Header) error {
	if l.email != "" {
		headers.Set("X-User-Email", l.email)
	}
	return nil
}
