package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/naistrix"
	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

const oidcLoginTimeout = 5 * time.Minute

var ErrNeedsOIDCLogin = errors.New("unauthenticated: please log in with `nais login -n`")

// NeedsLoginError is ErrNeedsOIDCLogin with the reason login is needed. Its message omits the reason,
// since the user logs in regardless; Cause holds it for verbose output.
type NeedsLoginError struct {
	Cause error
}

func (e *NeedsLoginError) Error() string { return ErrNeedsOIDCLogin.Error() }

func (e *NeedsLoginError) Unwrap() []error { return []error{ErrNeedsOIDCLogin, e.Cause} }

func needsLogin(format string, a ...any) error {
	return &NeedsLoginError{Cause: fmt.Errorf(format, a...)}
}

func OIDC(ctx context.Context) (*AuthenticatedUser, error) {
	user, err := getOIDCUser(ctx)
	if err != nil {
		return nil, err
	}

	client, err := newOidcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oidc config: %w", err)
	}

	idToken, err := client.ParseIDToken(ctx, user.IDToken)
	if errors.Is(err, jwt.TokenExpiredError()) {
		user, err = refreshOIDCUser(ctx, user)
		if err != nil {
			return nil, err
		}
		idToken, err = client.ParseIDToken(ctx, user.IDToken)
	}
	if err != nil {
		return nil, err
	}

	domain, ok := idToken.Domain()
	if !ok {
		return nil, needsLogin("missing primary_domain claim")
	}

	email, ok := idToken.Email()
	if !ok {
		return nil, needsLogin("missing email claim")
	}
	consoleHost, err := resolveConsoleHost(ctx, domain, user.ConsoleHost)
	if err != nil {
		return nil, err
	}

	ts := tokenSourceFunc(func() (*oauth2.Token, error) {
		user, err := getOIDCUser(ctx)
		if err != nil {
			return nil, err
		}

		return &user.Token, nil
	})

	return &AuthenticatedUser{
		consoleHost: consoleHost,
		domain:      domain,
		email:       email,
		ts:          oauth2.ReuseTokenSource(&user.Token, ts),
	}, nil
}

// OIDCLogin initiates the OpenID Connect authorization code flow to authenticate the user.
// Credentials are encrypted on disk with a key stored in the system keyring.
// See [AuthenticatedUser] for primitives that allows interacting with the Nais API on behalf of the authenticated user.
func OIDCLogin(ctx context.Context, out *naistrix.OutputWriter) error {
	ctx, timeoutCancel := context.WithTimeout(ctx, oidcLoginTimeout)
	defer timeoutCancel()

	client, err := newOidcClient(ctx)
	if err != nil {
		return err
	}

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()

	loginCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	srv := client.CallbackServer(loginCtx, cancel, verifier, state)

	// Listen before opening the browser, so the callback can't arrive first.
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("login failed: port 8865 is in use, possibly by another `nais login`: %w", err)
	}

	wg, ctx := errgroup.WithContext(loginCtx)
	wg.Go(func() error {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("login failed: callback server: %w", err)
		}
		return nil
	})

	url := client.oauth2.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	_ = urlopen.Open(url)

	out.Println("Your browser has been opened to visit:")
	out.Println()
	out.Println(url)
	out.Println()
	out.Println("If your browser didn't open, please copy the URL above and paste it in your browser's address bar")

	// context is canceled when the login callback handler completes once or when errgroup returns a non-nil error
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
	}

	if err := wg.Wait(); err != nil {
		return err
	}
	cause := context.Cause(loginCtx)
	if errors.Is(cause, context.DeadlineExceeded) {
		return fmt.Errorf("login timed out after %d minutes; run `nais login -n` to try again", int(oidcLoginTimeout.Minutes()))
	}
	if cause != nil && !errors.Is(cause, context.Canceled) {
		return cause
	}
	return nil
}

// OIDCLogout logs out of every config directory and ends the identity provider session.
func OIDCLogout(ctx context.Context, out *naistrix.OutputWriter) error {
	if err := deleteOIDCUser(ctx); err != nil {
		return err
	}

	client, err := newOidcClient(ctx)
	if err != nil {
		return fmt.Errorf("get oauth config: %w", err)
	}

	url := client.oidc.EndSessionEndpoint

	_ = urlopen.Open(url)

	out.Println("To complete logout, your browser has been opened to visit:")
	out.Println()
	out.Println(url)
	out.Println()
	out.Println("If your browser didn't open, please copy the URL above and paste it in your browser's address bar.")
	out.Println()

	return nil
}

type oidcClient struct {
	oauth2 *oauth2.Config
	oidc   *oidc.DiscoveryConfiguration
}

func newOidcClient(ctx context.Context) (*oidcClient, error) {
	clientID := "320114319427740585"
	if c := os.Getenv("NAIS_ZITADEL_CLIENT_ID"); c != "" {
		clientID = c
	}

	issuer := "https://auth.nais.io"
	if i := os.Getenv("NAIS_ZITADEL_DOMAIN"); i != "" {
		issuer = i
	}

	oidcCfg, err := oidcclient.Discover(ctx, issuer, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("discover oidc configuration from %q: %w", issuer, err)
	}

	oauth2Cfg := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "profile", "email", "offline_access", "urn:zitadel:iam:user:resourceowner"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oidcCfg.AuthorizationEndpoint,
			TokenURL: oidcCfg.TokenEndpoint,
		},
		RedirectURL: "http://localhost:8865/callback",
	}

	return &oidcClient{oauth2Cfg, oidcCfg}, nil
}

func (c *oidcClient) CallbackServer(ctx context.Context, cancel context.CancelCauseFunc, verifier, state string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "This login link is no longer valid. Return to your terminal and finish logging in, or run `nais login -n` again.", http.StatusBadRequest)
			return
		}

		var callbackErr error
		defer func() { cancel(callbackErr) }()

		fail := func(status int, err error) {
			callbackErr = fmt.Errorf("login failed: %w", err)
			http.Error(w, "Login failed. See your terminal for details.", status)
		}

		code := r.URL.Query().Get("code")

		tok, err := c.oauth2.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			fail(http.StatusInternalServerError, fmt.Errorf("exchange authorization code: %w", err))
			return
		}

		raw, ok := tok.Extra("id_token").(string)
		if !ok {
			fail(http.StatusInternalServerError, errors.New("invalid ID token: missing from token response"))
			return
		}

		idToken, err := c.ParseIDToken(ctx, raw)
		if err != nil {
			fail(http.StatusInternalServerError, fmt.Errorf("invalid ID token: %w", err))
			return
		}

		domain, ok := idToken.Domain()
		if !ok {
			fail(http.StatusUnauthorized, errors.New("invalid ID token: missing primary_domain claim"))
			return
		}
		consoleHost, err := resolveConsoleHost(ctx, domain, "")
		if err != nil {
			fail(http.StatusBadRequest, err)
			return
		}

		_, err = storeOIDCUser(ctx, tok, consoleHost)
		if err != nil {
			fail(http.StatusInternalServerError, fmt.Errorf("store credentials: %w", err))
			return
		}

		_, _ = fmt.Fprintln(w, "Successfully logged in! You can now close this window.")
	})

	return &http.Server{
		Addr:              "127.0.0.1:8865",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (c *oidcClient) ParseIDToken(ctx context.Context, token string) (*IDToken, error) {
	jwkSet, err := jwk.Fetch(ctx, c.oidc.JwksURI)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks from %q: %w", c.oidc.JwksURI, err)
	}

	j, err := jwt.ParseString(
		token,
		jwt.WithKeySet(jwkSet),
		jwt.WithIssuer(c.oidc.Issuer),
		jwt.WithAudience(c.oauth2.ClientID),
		jwt.WithClaimValue("email_verified", true),
		jwt.WithAcceptableSkew(10*time.Second),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, needsLogin("parse jwt: %w", err)
	}

	return &IDToken{j}, nil
}

type IDToken struct {
	jwt.Token
}

func (t *IDToken) Domain() (string, bool) {
	var domain string
	if err := t.Get("urn:zitadel:iam:user:resourceowner:primary_domain", &domain); err != nil {
		return "", false
	}

	return domain, true
}

func (t *IDToken) Email() (string, bool) {
	var email string
	if err := t.Get("email", &email); err != nil {
		return "", false
	}

	return email, true
}
