package naisapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/naistrix"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

var ErrNotAuthenticated = errors.New("not authenticated")

// AuthenticatedTokenUser represents the authenticated user.
// It provides primitives for interacting with the Nais API on behalf of the user.
// The primitives may return an [ErrNotAuthenticated] if the user has invalid or
// expired credentials, in which case the user must reauthenticate through [Login].
type AuthenticatedTokenUser struct {
	oauth2.TokenSource
	consoleHost string
	domain      string
}

type AuthenticatedUser interface {
	HTTPClient(ctx context.Context) *http.Client
	RoundTripper(base http.RoundTripper) http.RoundTripper
	SetAuthorizationHeader(headers http.Header) error
	ConsoleHost() string
	APIURL() string
	Domain() string
}

// GetAuthenticatedUser may return an [ErrNotAuthenticated] if the user has invalid or
// expired credentials, in which case the user must reauthenticate through [Login].
func GetAuthenticatedUser(ctx context.Context) (AuthenticatedUser, error) {
	if host := os.Getenv("NAIS_API_LOCAL_HOST"); host != "" {
		return &LocalhostUser{
			consoleHost: host,
			email:       os.Getenv("NAIS_API_LOCAL_EMAIL"),
		}, nil
	}

	githubTokenURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	githubTokenBearerToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if githubTokenURL != "" && githubTokenBearerToken != "" {
		// TODO: this is temporary; should be derived from the token somehow
		tenant := os.Getenv("NAIS_API_TENANT")
		if tenant == "" {
			return nil, fmt.Errorf("NAIS_API_TENANT must be set when using GitHub Actions authentication")
		}
		return &AuthenticatedTokenUser{
			TokenSource: oauth2.ReuseTokenSource(nil, &githubActionsTokenSource{
				ctx,
				githubTokenURL,
				githubTokenBearerToken,
			}),
			consoleHost: "console." + tenant + ".cloud.nais.io",
			domain:      "TODO",
		}, nil
	}

	secret, err := getUserSecret(ctx)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedTokenUser{
		TokenSource: oauth2.ReuseTokenSource(&secret.Token, &userTokenSource{ctx}),
		consoleHost: secret.ConsoleHost,
		domain:      secret.Domain,
	}, nil
}

func (a *AuthenticatedTokenUser) Domain() string {
	return a.domain
}

// ConsoleHost returns the console host of the authenticated user.
func (a *AuthenticatedTokenUser) ConsoleHost() string {
	return a.consoleHost
}

func (a *AuthenticatedTokenUser) APIURL() string {
	return fmt.Sprintf("https://%s/graphql", a.ConsoleHost())
}

// HTTPClient returns a [http.Client] configured with the user's access token.
func (a *AuthenticatedTokenUser) HTTPClient(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, a.TokenSource)
}

// RoundTripper returns a [http.RoundTripper] configured with the user's access token.
func (a *AuthenticatedTokenUser) RoundTripper(base http.RoundTripper) http.RoundTripper {
	return &oauth2.Transport{
		Base:   base,
		Source: a.TokenSource,
	}
}

// SetAuthorizationHeader sets the "Authorization" header with the user's access token.
func (a *AuthenticatedTokenUser) SetAuthorizationHeader(headers http.Header) error {
	tok, err := a.TokenSource.Token()
	if err != nil {
		return err
	}

	headers.Set("Authorization", "Bearer "+tok.AccessToken)
	return nil
}

// Login initiates the OAuth2 authorization code flow to authenticate the user.
// The user's secret is saved in the system keyring.
// See [AuthenticatedUser] for primitives that allows interacting with the Nais API on behalf of the authenticated user.
func Login(ctx context.Context, out *naistrix.OutputWriter) error {
	conf, oidcConfig, err := oauthConfig(ctx)
	if err != nil {
		return err
	}

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()
	ch := make(chan *oauth2.Token)

	go func() {
		if err := listenServer(ctx, conf, verifier, state, ch); err != nil {
			out.Println("Error starting server:", err)
			return
		}
	}()

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

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

	jwkSet, err := jwk.Fetch(ctx, oidcConfig.JwksURI)
	if err != nil {
		return fmt.Errorf("fetching jwks from %q: %w", oidcConfig.JwksURI, err)
	}

	idToken := tok.Extra("id_token").(string)
	j, err := jwt.ParseString(idToken,
		jwt.WithKeySet(jwkSet),
		jwt.WithValidate(true),
		jwt.WithIssuer(oauthIssuer()),
		jwt.WithAudience(oauthClientID()),
		jwt.WithClaimValue("email_verified", true),
	)
	if err != nil {
		return fmt.Errorf("parse jwt: %w", err)
	}

	var domain string
	err = j.Get("urn:zitadel:iam:user:resourceowner:primary_domain", &domain)
	if err != nil {
		return fmt.Errorf("getting primary_domain claim: %w", err)
	}

	tenantData, err := getTenantData(domain)
	if err != nil {
		return fmt.Errorf("getting tenant data: %w", err)
	}

	_, err = saveUserSecret(tok, domain, tenantData.ConsoleURL)
	if err != nil {
		return fmt.Errorf("saving token: %w", err)
	}
	return nil
}

// Logout deletes the user's secret from the system keyring and triggers logout at the identity provider.
func Logout(ctx context.Context, out *naistrix.OutputWriter) error {
	err := deleteSecret()
	if err != nil && !errors.Is(err, errSecretNotFound) {
		return fmt.Errorf("deleting user secret: %w", err)
	}

	_, oidcConfig, err := oauthConfig(ctx)
	if err != nil {
		return fmt.Errorf("getting oauth config: %w", err)
	}

	url := oidcConfig.EndSessionEndpoint

	_ = urlopen.Open(url)

	out.Println("To complete logout, your browser has been opened to visit:")
	out.Println()
	out.Println(url)
	out.Println()
	out.Println("If your browser didn't open, please copy the URL above and paste it in your browser's address bar.")
	out.Println()

	return nil
}

// userSecret defines the data to marshal to and from the system keyring.
type userSecret struct {
	oauth2.Token
	IDToken     string `json:"id_token"`
	ConsoleHost string `json:"console_host"`
	Domain      string `json:"domain"`
}

type tenantData struct {
	ConsoleURL string `json:"consoleUrl"`
}

type userTokenSource struct {
	ctx context.Context
}

func (u *userTokenSource) Token() (*oauth2.Token, error) {
	secret, err := getUserSecret(u.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user secret: %w", err)
	}

	return &secret.Token, nil
}

type githubActionsTokenSource struct {
	ctx          context.Context
	requestURL   string
	requestToken string
}

func (g *githubActionsTokenSource) Token() (*oauth2.Token, error) {
	req, err := http.NewRequestWithContext(g.ctx, http.MethodGet, g.requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Add("audience", "api.nais.io")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", "bearer "+g.requestToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var tokenResponse struct {
		Token string `json:"value"`
	}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return &oauth2.Token{
		AccessToken: tokenResponse.Token,
	}, nil
}

func getUserSecret(ctx context.Context) (*userSecret, error) {
	secretData, err := getSecret()
	if err != nil {
		if errors.Is(err, errSecretNotFound) {
			return nil, ErrNotAuthenticated
		}
		return nil, fmt.Errorf("getting user secret: %w", err)
	}

	var sec userSecret
	err = json.Unmarshal([]byte(secretData), &sec)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling data from keyring")
	}

	if !sec.Valid() {
		return refreshUserToken(ctx, &sec)
	}

	return &sec, nil
}

func saveUserSecret(tok *oauth2.Token, domain, consoleURL string) (*userSecret, error) {
	sec := &userSecret{
		Token:       *tok,
		IDToken:     tok.Extra("id_token").(string),
		ConsoleHost: consoleURL,
		Domain:      domain,
	}
	secret, err := json.Marshal(sec)
	if err != nil {
		return nil, fmt.Errorf("marshalling token: %w", err)
	}

	err = setSecret(string(secret))
	if err != nil {
		return nil, fmt.Errorf("setting user secret: %w", err)
	}
	return sec, nil
}

func refreshUserToken(ctx context.Context, sec *userSecret) (*userSecret, error) {
	cfg, _, err := oauthConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting oauth config: %w", err)
	}

	tok, err := cfg.TokenSource(ctx, &sec.Token).Token()
	if err != nil {
		return nil, ErrNotAuthenticated
	}

	return saveUserSecret(tok, sec.Domain, sec.ConsoleHost)
}

func getTenantData(domain string) (*tenantData, error) {
	u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", domain)
	res, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("getting %q: %w", u, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting %q: %v", u, res.Status)
	}

	var data tenantData
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding: %w", err)
	}

	return &data, nil
}

func oauthConfig(ctx context.Context) (*oauth2.Config, *oidc.DiscoveryConfiguration, error) {
	oidcConfig, err := client.Discover(ctx, oauthIssuer(), http.DefaultClient)
	if err != nil {
		return nil, nil, fmt.Errorf("discover oidc configuration from %q: %w", oauthIssuer(), err)
	}

	conf := &oauth2.Config{
		ClientID: oauthClientID(),
		Scopes:   []string{"openid", "profile", "email", "offline_access", "urn:zitadel:iam:user:resourceowner"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oidcConfig.AuthorizationEndpoint,
			TokenURL: oidcConfig.TokenEndpoint,
		},
		RedirectURL: "http://localhost:8865/callback",
	}
	return conf, oidcConfig, nil
}

func oauthIssuer() string {
	if issuer := os.Getenv("NAIS_ZITADEL_DOMAIN"); issuer != "" {
		return issuer
	}
	return "https://auth.nais.io"
}

func oauthClientID() string {
	if clientID := os.Getenv("NAIS_ZITADEL_CLIENT_ID"); clientID != "" {
		return clientID
	}
	return "320114319427740585"
}

func listenServer(ctx context.Context, cfg *oauth2.Config, verifier, state string, ch chan *oauth2.Token) error {
	srv := &http.Server{Addr: ":8865"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State did not match", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")

		tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintln(w, "Success! You can now close this window.")

		ch <- tok
	})

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
