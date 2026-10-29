package auth

import (
	"context"
	"errors"
	"fmt"
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

var ErrNeedsOIDCLogin = errors.New("unauthenticated: please log in with `nais login -n`")

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
	if err != nil {
		return nil, ErrNeedsOIDCLogin
	}

	domain, ok := idToken.Domain()
	if !ok {
		return nil, fmt.Errorf("%w: missing primary_domain claim", ErrNeedsOIDCLogin)
	}

	email, ok := idToken.Email()
	if !ok {
		return nil, fmt.Errorf("%w: missing email claim", ErrNeedsOIDCLogin)
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
	client, err := newOidcClient(ctx)
	if err != nil {
		return err
	}

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()

	loginCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	srv := client.CallbackServer(loginCtx, cancel, verifier, state)

	wg, ctx := errgroup.WithContext(loginCtx)
	wg.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("error starting login callback server: %w", err)
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
	_ = srv.Shutdown(context.Background())
	if err := wg.Wait(); err != nil {
		return err
	}
	if cause := context.Cause(loginCtx); cause != nil && !errors.Is(cause, context.Canceled) {
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
		var callbackErr error
		defer func() { cancel(callbackErr) }()

		if r.URL.Query().Get("state") != state {
			callbackErr = errors.New("state did not match")
			http.Error(w, callbackErr.Error(), http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")

		tok, err := c.oauth2.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			callbackErr = fmt.Errorf("failed to exchange token: %w", err)
			http.Error(w, callbackErr.Error(), http.StatusInternalServerError)
			return
		}

		raw, ok := tok.Extra("id_token").(string)
		if !ok {
			callbackErr = errors.New("missing id_token in token response")
			http.Error(w, callbackErr.Error(), http.StatusInternalServerError)
			return
		}

		idToken, err := c.ParseIDToken(ctx, raw)
		if err != nil {
			callbackErr = fmt.Errorf("failed to parse ID token: %w", err)
			http.Error(w, callbackErr.Error(), http.StatusInternalServerError)
			return
		}

		domain, ok := idToken.Domain()
		if !ok {
			callbackErr = errors.New("missing primary_domain claim in ID token")
			http.Error(w, callbackErr.Error(), http.StatusUnauthorized)
			return
		}
		consoleHost, err := resolveConsoleHost(ctx, domain, "")
		if err != nil {
			callbackErr = err
			http.Error(w, callbackErr.Error(), http.StatusBadRequest)
			return
		}

		_, err = storeOIDCUser(ctx, tok, consoleHost)
		if err != nil {
			callbackErr = fmt.Errorf("failed to store user data: %w", err)
			http.Error(w, callbackErr.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintln(w, "Successfully logged in! You can now close this window.")
	})

	return &http.Server{
		Addr:    ":8865",
		Handler: mux,
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
		return nil, fmt.Errorf("parse jwt: %w", err)
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
