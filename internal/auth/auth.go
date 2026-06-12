package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
)

// GetEnvironmentOIDCIssuer returns the OIDC issuer URL for workload identity
// tokens in the given environment.
func GetEnvironmentOIDCIssuer(ctx context.Context, env string) (*url.URL, error) {
	_ = `# @genqlient
		query EnvironmentOIDCIssuer($name: String!) {
			environment(name: $name) {
				oidcIssuerURL
			}
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := gql.EnvironmentOIDCIssuer(ctx, client, env)
	if err != nil {
		return nil, err
	}

	if resp.Environment.OidcIssuerURL == "" {
		return nil, fmt.Errorf("environment %q does not support workload identity (no OIDC issuer URL)", env)
	}

	issuer, err := url.Parse(resp.Environment.OidcIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("parsing OIDC issuer URL %q for environment %q: %w", resp.Environment.OidcIssuerURL, env, err)
	}

	if issuer.Scheme != "https" || issuer.Host == "" || issuer.RawQuery != "" || issuer.Fragment != "" {
		return nil, fmt.Errorf("invalid OIDC issuer URL %q for environment %q: expected an absolute https URL without query or fragment", resp.Environment.OidcIssuerURL, env)
	}

	return issuer, nil
}

// FetchOIDCDiscoveryDocument resolves the OIDC discovery document for the given
// issuer URL (by appending /.well-known/openid-configuration) and returns the
// decoded JSON document.
func FetchOIDCDiscoveryDocument(ctx context.Context, issuer *url.URL) (any, error) {
	discoveryURL := issuer.JoinPath(".well-known", "openid-configuration").String()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching OIDC discovery document: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching OIDC discovery document from %s: unexpected status %s: %s", discoveryURL, resp.Status, body)
	}

	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("OIDC discovery document from %s is not valid JSON: %w", discoveryURL, err)
	}

	return doc, nil
}
