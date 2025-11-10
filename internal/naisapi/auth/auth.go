package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

var ErrNotAuthenticated = errors.New("not authenticated")

type AuthenticatedUser struct {
	oauth2.TokenSource
	consoleHost string
	domain      string
}

func (a *AuthenticatedUser) Domain() string {
	return a.domain
}

func (a *AuthenticatedUser) ConsoleHost() string {
	return a.consoleHost
}

func (a *AuthenticatedUser) APIURL() string {
	return fmt.Sprintf("https://%s/graphql", a.ConsoleHost())
}

func (a *AuthenticatedUser) HTTPClient(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, a.TokenSource)
}

func (a *AuthenticatedUser) RoundTripper(base http.RoundTripper) http.RoundTripper {
	return &oauth2.Transport{
		Base:   base,
		Source: a.TokenSource,
	}
}

func (a *AuthenticatedUser) SetAuthorizationHeader(headers http.Header) error {
	tok, err := a.TokenSource.Token()
	if err != nil {
		return err
	}

	headers.Set("Authorization", "Bearer "+tok.AccessToken)
	return nil
}

func (a *AuthenticatedUser) GetTokenSource() oauth2.TokenSource {
	return a.TokenSource
}
