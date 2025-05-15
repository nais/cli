package nais

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/nais/cli/internal/urlopen"
	"golang.org/x/oauth2"
)

// FIXME: should be overridable by env vars
const zitadelDomain = "https://auth.nais.io"
const zitadelClientID = "320114319427740585"

func Login(ctx context.Context) error {
	issuer := os.Getenv("NAIS_ZITADEL_DOMAIN")
	if issuer == "" {
		issuer = zitadelDomain
	}

	clientID := os.Getenv("NAIS_ZITADEL_CLIENT_ID")
	if clientID == "" {
		clientID = zitadelClientID
	}

	conf := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "profile", "email", "urn:zitadel:iam:user:resourceowner"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuer + "/oauth/v2/authorize",
			TokenURL: issuer + "/oauth/v2/token",
		},
		RedirectURL: "http://localhost:8865/callback",
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
		fmt.Printf("Token: %v\n", tok)
	}

	idToken := tok.Extra("id_token").(string)

	// FIXME: verify the token's signature and validate its standard claims, check `email_verified` claim
	j, err := jwt.ParseString(idToken, jwt.WithVerify(false))
	if err != nil {
		return fmt.Errorf("parse jwt: %w", err)
	}

	domainClaim, ok := j.Get("urn:zitadel:iam:user:resourceowner:primary_domain")
	if !ok {
		return fmt.Errorf("missing claim 'urn:zitadel:iam:user:resourceowner:primary_domain' in jwt")
	}

	domain, ok := domainClaim.(string)
	if !ok {
		return fmt.Errorf("claim 'urn:zitadel:iam:user:resourceowner:primary_domain' is not a string")
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

	fmt.Printf("Tenant data: %s\n", tenantData.ConsoleURL)

	body := `{"query":"query Teams {\n  me {\n    ... on User {\n      id\n      email\n    }\n  }\n}","operationName":"Teams"}`
	req, err := http.NewRequest("POST", "http://localhost:3000/graphql", strings.NewReader(body))
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
	fmt.Println(tok.AccessToken)
	fmt.Println("----")
	io.Copy(os.Stdout, res.Body)
	fmt.Println("----")

	// TODO:
	// - store tokens in keyring
	// - use the access token to call the API
	// - use the refresh token to get a new access token as necessary
	return nil
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

	srv.ListenAndServe()
}
