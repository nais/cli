package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/nais/cli/internal/keyring"
	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/naistrix"
	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

var ErrNeedsOIDCLogin = errors.New("unauthenticated: please log in with `nais auth login -n`")

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

	ts := tokenSourceFunc(func() (*oauth2.Token, error) {
		user, err := getOIDCUser(ctx)
		if err != nil {
			return nil, err
		}

		return &user.Token, nil
	})

	return &AuthenticatedUser{
		consoleHost: user.ConsoleHost,
		domain:      domain,
		email:       email,
		ts:          oauth2.ReuseTokenSourceWithExpiry(&user.Token, ts, 30*time.Second),
	}, nil
}

// OIDCLogin initiates the OpenID Connect authorization code flow to authenticate the user.
// The user's secret is saved in the system keyring.
// See [AuthenticatedUser] for primitives that allows interacting with the Nais API on behalf of the authenticated user.
func OIDCLogin(ctx context.Context, out *naistrix.OutputWriter) error {
	client, err := newOidcClient(ctx)
	if err != nil {
		return err
	}

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()

	ctx, cancel := context.WithCancel(ctx)
	srv := client.CallbackServer(ctx, cancel, verifier, state)

	wg, ctx := errgroup.WithContext(ctx)
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
	return wg.Wait()
}

// OIDCLogout deletes the user's secret from the system keyring and triggers logout at the identity provider.
func OIDCLogout(ctx context.Context, out *naistrix.OutputWriter) error {
	err := keyring.Delete()
	if err != nil && !errors.Is(err, keyring.ErrSecretNotFound) {
		return fmt.Errorf("delete user secret: %w", err)
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

func (c *oidcClient) CallbackServer(ctx context.Context, cancel context.CancelFunc, verifier, state string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer cancel()

		if r.URL.Query().Get("state") != state {
			http.Error(w, "State did not match", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")

		tok, err := c.oauth2.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		raw, ok := tok.Extra("id_token").(string)
		if !ok {
			http.Error(w, "Missing id_token in token response", http.StatusInternalServerError)
			return
		}

		idToken, err := c.ParseIDToken(ctx, raw)
		if err != nil {
			http.Error(w, "Failed to parse ID token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		domain, ok := idToken.Domain()
		if !ok {
			http.Error(w, "Missing primary_domain claim in ID token", http.StatusUnauthorized)
			return
		}
		if domain == "nais.io" {
			// TODO: we should probably support this at some point
			http.Error(w, "Cannot login with `@nais.io` user; please use a tenant-specific account", http.StatusBadRequest)
			return
		}

		type tenantData struct {
			ConsoleURL string `json:"consoleUrl"`
		}
		u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", domain)
		res, err := http.Get(u)
		if err != nil {
			http.Error(w, "Failed to get tenant data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = res.Body.Close()
		}()
		if res.StatusCode == http.StatusNotFound {
			http.Error(w, fmt.Sprintf("No tenant data found for domain %q", domain), http.StatusBadRequest)
			return
		}
		if res.StatusCode != http.StatusOK {
			http.Error(w, "Failed to get tenant data: unexpected status code "+res.Status, http.StatusInternalServerError)
			return
		}

		var tenant tenantData
		err = json.NewDecoder(res.Body).Decode(&tenant)
		if err != nil {
			http.Error(w, "Failed to decode tenant data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = storeOIDCUser(tok, tenant.ConsoleURL)
		if err != nil {
			http.Error(w, "Failed to store user data: "+err.Error(), http.StatusInternalServerError)
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

	j, err := jwt.ParseString(token,
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
