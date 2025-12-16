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
	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/naistrix"
	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
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
	ch := make(chan *oauth2.Token)

	go func() {
		if err := client.CallbackServer(ctx, verifier, state, ch); err != nil {
			out.Println("Error starting server:", err)
			return
		}
	}()

	url := client.oauth2.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

	_ = urlopen.Open(url)

	out.Println("Your browser has been opened to visit:")
	out.Println()
	out.Println(url)
	out.Println()
	out.Println("If your browser didn't open, please copy the URL above and paste it in your browser's address bar")

	var tok *oauth2.Token
	select {
	case <-ctx.Done():
		return nil
	case tok = <-ch:
	}

	raw, ok := tok.Extra("id_token").(string)
	if !ok {
		return fmt.Errorf("missing id_token in token response")
	}

	idToken, err := client.ParseIDToken(ctx, raw)
	if err != nil {
		return fmt.Errorf("parse id token: %w", err)
	}

	domain, ok := idToken.Domain()
	if !ok {
		return fmt.Errorf("missing primary_domain claim")
	}

	type tenantData struct {
		ConsoleURL string `json:"consoleUrl"`
	}
	u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", domain)
	res, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("get %q: %w", u, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("get %q: unexpected status code: %q", u, res.Status)
	}

	var tenant tenantData
	err = json.NewDecoder(res.Body).Decode(&tenant)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	_, err = storeOIDCUser(tok, tenant.ConsoleURL)
	if err != nil {
		return fmt.Errorf("store oidcUser: %w", err)
	}

	return nil
}

// OIDCLogout deletes the user's secret from the system keyring and triggers logout at the identity provider.
func OIDCLogout(ctx context.Context, out *naistrix.OutputWriter) error {
	err := deleteKeyringSecret()
	if err != nil && !errors.Is(err, errSecretNotFound) {
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

// oidcUser defines the data to marshal to and from the system keyring.
type oidcUser struct {
	oauth2.Token
	IDToken     string `json:"id_token"`
	ConsoleHost string `json:"console_host"`
}

func (u *oidcUser) Refresh(ctx context.Context) (*oidcUser, error) {
	client, err := newOidcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oauth config: %w", err)
	}

	tok, err := client.oauth2.TokenSource(ctx, &u.Token).Token()
	if err != nil {
		return nil, ErrNeedsOIDCLogin
	}

	user, err := storeOIDCUser(tok, u.ConsoleHost)
	if err != nil {
		return nil, fmt.Errorf("%w: %+v", ErrNeedsOIDCLogin, err)
	}

	return user, nil
}

func getOIDCUser(ctx context.Context) (*oidcUser, error) {
	secret, err := getKeyringSecret()
	if err != nil {
		if errors.Is(err, errSecretNotFound) {
			return nil, ErrNeedsOIDCLogin
		}
		return nil, fmt.Errorf("get oidc user: %w", err)
	}

	var user oidcUser
	err = json.Unmarshal([]byte(secret), &user)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data from keyring")
	}

	if !user.Valid() {
		return user.Refresh(ctx)
	}

	return &user, nil
}

func storeOIDCUser(tok *oauth2.Token, consoleURL string) (*oidcUser, error) {
	idToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token")
	}

	sec := &oidcUser{
		Token:       *tok,
		IDToken:     idToken,
		ConsoleHost: consoleURL,
	}

	secret, err := json.Marshal(sec)
	if err != nil {
		return nil, fmt.Errorf("marshal token: %w", err)
	}

	err = setKeyringSecret(string(secret))
	if err != nil {
		return nil, fmt.Errorf("set keyring secret: %w", err)
	}

	return sec, nil
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

func (c *oidcClient) CallbackServer(ctx context.Context, verifier, state string, ch chan *oauth2.Token) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
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

		_, _ = fmt.Fprintln(w, "Success! You can now close this window.")

		ch <- tok
	})

	srv := &http.Server{
		Addr:    ":8865",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
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
