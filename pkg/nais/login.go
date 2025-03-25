package nais

import (
	"context"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/oauth2"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sts/v1"
)

const zitadelDomain = "https://login-test.nais.io"
const zitadelClientID = "312750485188752940"

func Login(ctx context.Context) error {
	conf := &oauth2.Config{
		ClientID: zitadelClientID,
		Scopes:   []string{"openid", "profile", "email", "urn:zitadel:iam:user:resourceowner"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  zitadelDomain + "/oauth/v2/authorize",
			TokenURL: zitadelDomain + "/oauth/v2/token",
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
	fmt.Println("Visit the URL for the auth dialog:\n", url)

	var tok *oauth2.Token
	select {
	case <-ctx.Done():
		return nil
	case tok = <-ch:
		fmt.Printf("Token: %v\n", tok)
	}

	idToken := tok.Extra("id_token").(string)

	// Parse access token as jwt
	j, err := jwt.ParseString(idToken, jwt.WithVerify(false))
	if err != nil {
		return fmt.Errorf("parse jwt: %w", err)
	}
	orgID, ok := j.Get("urn:zitadel:iam:user:resourceowner:id")
	if !ok {
		return fmt.Errorf("missing claim 'urn:zitadel:iam:user:resourceowner:id' in jwt")
	}

	stsClient, err := sts.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return err
	}
	stsResp, err := stsClient.V1.Token(&sts.GoogleIdentityStsV1ExchangeTokenRequest{
		Audience:           "//iam.googleapis.com/locations/global/workforcePools/zitadel-nais/providers/" + orgID.(string),
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		Scope:              "https://www.googleapis.com/auth/cloud-platform",
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:id_token",
		SubjectToken:       idToken,
	}).Context(ctx).Do()
	if err != nil {
		return err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: stsResp.AccessToken})
	svc, err := container.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return err
	}

	call := svc.Projects.Locations.Clusters.List("projects/nais-dev-cdea/locations/-")
	response, err := call.Do()
	if err != nil {
		return err
	}

	spew.Dump(response)

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
