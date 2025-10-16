package naisapi

import (
	"context"
	"fmt"
	"net/http"
)

type LocalhostUser struct {
	consoleHost string
	email       string
}

func (l *LocalhostUser) Domain() string {
	return "example.com"
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

type LocalhostRoundtripper struct {
	user *LocalhostUser
	base http.RoundTripper
}

func (l *LocalhostRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	_ = l.user.SetAuthorizationHeader(r.Header)
	return l.base.RoundTrip(r)
}
