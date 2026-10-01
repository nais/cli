package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

func TestOIDCWithStoredSession(t *testing.T) {
	keyring.MockInit()
	path := filepath.Join(t.TempDir(), "config.yaml")
	ConfigFilePath = &path
	t.Cleanup(func() { ConfigFilePath = nil })

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := jwk.Import(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := publicKey.Set(jwk.KeyIDKey, "test"); err != nil {
		t.Fatal(err)
	}
	if err := publicKey.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		t.Fatal(err)
	}
	keys := jwk.NewSet()
	if err := keys.AddKey(publicKey); err != nil {
		t.Fatal(err)
	}

	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer": issuer, "jwks_uri": issuer + "/keys",
				"authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token",
			})
		case "/keys":
			_ = json.NewEncoder(w).Encode(keys)
		default:
			t.Errorf("unexpected request: %s", r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	issuer = server.URL
	t.Setenv("NAIS_ZITADEL_DOMAIN", issuer)

	idToken := jwt.New()
	for name, value := range map[string]any{
		"iss": issuer, "aud": "320114319427740585", "exp": time.Now().Add(time.Hour),
		"email": "user@example.com", "email_verified": true,
		"urn:zitadel:iam:user:resourceowner:primary_domain": "example.com",
	} {
		if err := idToken.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	headers := jws.NewHeaders()
	if err := headers.Set(jws.KeyIDKey, "test"); err != nil {
		t.Fatal(err)
	}
	signed, err := jwt.Sign(idToken, jwt.WithKey(jwa.RS256(), key, jws.WithProtectedHeaders(headers)))
	if err != nil {
		t.Fatal(err)
	}
	_, err = storeOIDCUser((&oauth2.Token{
		AccessToken: "access-token", Expiry: time.Now().Add(time.Hour),
	}).WithExtra(map[string]any{"id_token": string(signed)}), "console.example.com")
	if err != nil {
		t.Fatal(err)
	}

	user, err := OIDC(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := user.Email(), "user@example.com"; got != want {
		t.Errorf("Email() = %q, want %q", got, want)
	}
}
