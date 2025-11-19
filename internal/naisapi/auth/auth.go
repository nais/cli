package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type AuthenticatedUser struct {
	consoleHost string
	domain      string
	email       string
	ts          oauth2.TokenSource
}

func (a *AuthenticatedUser) AccessToken() (string, error) {
	tok, err := a.ts.Token()
	if err != nil {
		return "", err
	}

	return tok.AccessToken, nil
}

func (a *AuthenticatedUser) APIURL() string {
	return fmt.Sprintf("https://%s/graphql", a.ConsoleHost())
}

func (a *AuthenticatedUser) ConsoleHost() string {
	return a.consoleHost
}

func (a *AuthenticatedUser) Domain() string {
	return a.domain
}

func (a *AuthenticatedUser) Email() string {
	return a.email
}

func (a *AuthenticatedUser) HTTPClient(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, a.ts)
}

func (a *AuthenticatedUser) RoundTripper(base http.RoundTripper) http.RoundTripper {
	return &oauth2.Transport{
		Base:   base,
		Source: a.ts,
	}
}

func (a *AuthenticatedUser) SetAuthorizationHeader(headers http.Header) error {
	tok, err := a.ts.Token()
	if err != nil {
		return err
	}

	headers.Set("Authorization", "Bearer "+tok.AccessToken)
	return nil
}

type tokenSourceFunc func() (*oauth2.Token, error)

func (t tokenSourceFunc) Token() (*oauth2.Token, error) {
	return t()
}

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}
