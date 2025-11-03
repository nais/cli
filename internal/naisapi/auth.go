package naisapi

import (
	"context"
	"net/http"

	"github.com/nais/cli/internal/naisapi/auth"
	"github.com/nais/naistrix"
)

var ErrNotAuthenticated = auth.ErrNotAuthenticated

// AuthenticatedUser represents the authenticated user.
// It provides primitives for interacting with the Nais API on behalf of the user.
// The primitives may return an [ErrNotAuthenticated] if the user has invalid or
// expired credentials, in which case the user must reauthenticate through [Login].
type AuthenticatedUser interface {
	// HTTPClient returns a [http.Client] configured with the user's credentials.
	HTTPClient(ctx context.Context) *http.Client
	// RoundTripper returns a [http.RoundTripper] configured with the user's credentials.
	RoundTripper(base http.RoundTripper) http.RoundTripper
	// SetAuthorizationHeader sets the "Authorization" header with the user's credentials.
	SetAuthorizationHeader(headers http.Header) error
	ConsoleHost() string
	APIURL() string
	Domain() string
}

// GetAuthenticatedUser may return an [ErrNotAuthenticated] if the user has invalid or
// expired credentials, in which case the user must reauthenticate through [Login].
func GetAuthenticatedUser(ctx context.Context) (AuthenticatedUser, error) {
	local, ok := auth.Localhost()
	if ok {
		return local, nil
	}

	githubActions, ok, err := auth.GithubActions(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		return githubActions, nil
	}

	return auth.OIDC(ctx)
}

// Login logs the user in to allow authenticated requests to the Nais API.
func Login(ctx context.Context, out *naistrix.OutputWriter) error {
	return auth.OIDCLogin(ctx, out)
}

// Logout logs the user out so that they can no longer make authenticated requests to the Nais API.
func Logout(ctx context.Context, out *naistrix.OutputWriter) error {
	return auth.OIDCLogout(ctx, out)
}
