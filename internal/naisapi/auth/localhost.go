package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func Localhost() (*LocalhostUser, bool) {
	host := os.Getenv("NAIS_API_LOCAL_HOST")
	if host == "" {
		return nil, false
	}

	return &LocalhostUser{
		consoleHost: host,
		email:       os.Getenv("NAIS_API_LOCAL_EMAIL"),
	}, true
}

type LocalhostUser struct {
	consoleHost string
	email       string
}

type LocalhostTokenSource struct{}

func (l *LocalhostTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: "not-really-an-access-token",
	}, nil
}

func (l *LocalhostUser) Domain() string {
	return "example.com"
}

func (l *LocalhostUser) Email() string {
	return l.email
}

func (l *LocalhostUser) ConsoleHost() string {
	return l.consoleHost
}

func (l *LocalhostUser) APIURL() string {
	return fmt.Sprintf("http://%s/graphql", l.ConsoleHost())
}

func (l *LocalhostUser) HTTPClient(ctx context.Context) *http.Client {
	return &http.Client{
		Transport: l.RoundTripper(http.DefaultTransport),
	}
}

func (l *LocalhostUser) RoundTripper(base http.RoundTripper) http.RoundTripper {
	return &LocalhostRoundtripper{
		user: l,
		base: base,
	}
}

func (l *LocalhostUser) SetAuthorizationHeader(headers http.Header) error {
	if l.email != "" {
		headers.Set("X-User-Email", l.email)
	}
	return nil
}

func (l *LocalhostUser) GetTokenSource() oauth2.TokenSource {
	return &LocalhostTokenSource{}
}

type LocalhostRoundtripper struct {
	user *LocalhostUser
	base http.RoundTripper
}

func (l *LocalhostRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	_ = l.user.SetAuthorizationHeader(r.Header)
	return l.base.RoundTrip(r)
}
