package nais

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/nais/cli/internal/urlopen"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
)

const (
	zitadelDomain   = "https://auth.nais.io"
	zitadelClientID = "320114319427740585"
)

type secret struct {
	oauth2.Token
	IDToken     string `json:"id_token"`
	ConsoleHost string `json:"console_host"`
}

func Login(ctx context.Context) error {
	conf, oidcClient, err := oauthConfig(ctx)
	if err != nil {
		return err
	}

	state := uuid.New().String()
	verifier := oauth2.GenerateVerifier()
	ch := make(chan *oauth2.Token)

	go listenServer(ctx, conf, verifier, state, ch)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Println("Your browser has been opened to visit:")
	fmt.Println()
	fmt.Println(url)
	fmt.Println()
	fmt.Println("If your browser doesn't open, please visit the URL above")

	_ = urlopen.Open(url)

	var tok *oauth2.Token
	select {
	case <-ctx.Done():
		return nil
	case tok = <-ch:
	}

	idToken := tok.Extra("id_token").(string)

	set, err := jwk.Fetch(ctx, oidcClient.JwksURI)
	if err != nil {
		return fmt.Errorf("fetching jwks from %q: %w", oidcClient.JwksURI, err)
	}

	// FIXME: verify the token's signature and validate its standard claims, check `email_verified` claim
	j, err := jwt.ParseString(idToken, jwt.WithKeySet(set))
	if err != nil {
		return fmt.Errorf("parse jwt: %w", err)
	}

	err = jwt.Validate(
		j,
		jwt.WithIssuer(oauthIssuer()),
		jwt.WithAudience(oauthClientID()),
		jwt.WithClaimValue("email_verified", true),
	)
	if err != nil {
		return fmt.Errorf("validating jwt: %w", err)
	}

	var domain string
	err = j.Get("urn:zitadel:iam:user:resourceowner:primary_domain", &domain)
	if err != nil {
		return fmt.Errorf("getting primary_domain claim: %w", err)
	}

	u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", domain)
	res, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("failed to get tenant data at %q: %w", u, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get tenant data at %q: %v", u, res.Status)
	}

	var tenantData struct {
		ConsoleURL string `json:"consoleUrl"`
	}
	err = json.NewDecoder(res.Body).Decode(&tenantData)
	if err != nil {
		return fmt.Errorf("decode tenant data: %w", err)
	}

	if _, err := saveToken(tok, tenantData.ConsoleURL); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	// use the access token to call the API, should be moved to appropriate package
	consoleURL := fmt.Sprintf("https://%s/graphql", tenantData.ConsoleURL)
	fmt.Printf("Querying %q\n", consoleURL)

	body := `{"query":"query Teams {\n  me {\n    ... on User {\n      id\n      email\n    }\n  }\n}","operationName":"Teams"}`
	req, err := http.NewRequest("POST", consoleURL, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	tok.SetAuthHeader(req)
	req.Header.Set("Accept-content", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}
	defer res.Body.Close()
	fmt.Println("----")
	data := make(map[string]any)
	err = json.NewDecoder(res.Body).Decode(&data)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
	fmt.Println()
	fmt.Println("----")

	return nil
}

func saveToken(tok *oauth2.Token, consoleURL string) (*secret, error) {
	sec := &secret{
		Token:       *tok,
		IDToken:     tok.Extra("id_token").(string),
		ConsoleHost: consoleURL,
	}
	secret, err := json.Marshal(sec)
	if err != nil {
		return nil, fmt.Errorf("marshalling token: %w", err)
	}

	err = setSecret("nais-user", string(secret))
	if err != nil {
		return nil, fmt.Errorf("setting secret: %w", err)
	}
	return sec, nil
}

func oauthConfig(ctx context.Context) (*oauth2.Config, *oidc.DiscoveryConfiguration, error) {
	oidcClient, err := client.Discover(ctx, oauthIssuer(), http.DefaultClient)
	if err != nil {
		return nil, nil, fmt.Errorf("discover oidc configuration from %q: %w", oauthIssuer(), err)
	}

	conf := &oauth2.Config{
		ClientID: oauthClientID(),
		Scopes:   []string{"openid", "profile", "email", "urn:zitadel:iam:user:resourceowner"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  oidcClient.AuthorizationEndpoint,
			TokenURL: oidcClient.TokenEndpoint,
		},
		RedirectURL: "http://localhost:8865/callback",
	}
	return conf, oidcClient, nil
}

func oauthIssuer() string {
	issuer := os.Getenv("NAIS_ZITADEL_DOMAIN")
	if issuer == "" {
		return zitadelDomain
	}
	return issuer
}

func oauthClientID() string {
	clientID := os.Getenv("NAIS_ZITADEL_CLIENT_ID")
	if clientID == "" {
		return zitadelClientID
	}
	return clientID
}

// AuthenticatedHTTPClient returns a HTTP client configured with the user's access token.
// Fetch and refresh tokens from as necessary.
func AuthenticatedHTTPClient() (*http.Client, error) {
	panic("not implemented")
}

func listenServer(ctx context.Context, cfg *oauth2.Config, verifier, state string, ch chan *oauth2.Token) {
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

		fmt.Fprintln(w, "Success! You can now close this window.")

		ch <- tok
	})

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprint(os.Stderr, "Errored while starting server: ", err)
	}
}

func getUserToken(ctx context.Context) (*secret, error) {
	secretData, err := getSecret("nais-user")
	if err != nil {
		return nil, fmt.Errorf("getting secret: %w", err)
	}
	var s secret
	err = json.Unmarshal([]byte(secretData), &s)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling data from keyring")
	}

	if !s.Valid() {
		// Try to refresh the token
		return refreshToken(ctx, s)
	}

	return &s, nil
}

func refreshToken(ctx context.Context, s secret) (*secret, error) {
	cfg, _, err := oauthConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting oauth config: %w", err)
	}
	tok, err := cfg.TokenSource(ctx, &s.Token).Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	return saveToken(tok, s.ConsoleHost)
}
