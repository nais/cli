package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
	"golang.org/x/oauth2"
)

func GithubActions(ctx context.Context) (*AuthenticatedUser, bool, error) {
	u := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	tok := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if u == "" || tok == "" {
		return nil, false, nil
	}

	tenant, err := resolveGithubTenant(ctx)
	if err != nil {
		return nil, false, err
	}
	domain := "TODO"

	ts := githubTokenSource(ctx, u, tok)
	token, err := ts.Token()
	if err != nil {
		return nil, false, fmt.Errorf("getting token from github actions: %w", err)
	}

	return &AuthenticatedUser{
		consoleHost: "console." + tenant + ".cloud.nais.io",
		domain:      domain,
		ts:          oauth2.ReuseTokenSourceWithExpiry(token, ts, 30*time.Second),
	}, true, nil
}

func resolveGithubTenant(ctx context.Context) (string, error) {
	// Allow explicit override via environment variable
	if tenant := os.Getenv("NAIS_API_TENANT"); tenant != "" {
		return tenant, nil
	}

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	if owner == "" {
		return "", fmt.Errorf("GITHUB_REPOSITORY_OWNER must be set when using GitHub Actions authentication")
	}

	u := fmt.Sprintf("https://storage.googleapis.com/github-deploy-data/%s.json", owner)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("creating deploy data request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req) // #nosec G704 -- URL is constructed from a fixed GCS base URL with owner segment from GITHUB_REPOSITORY_OWNER; SSRF risk is acceptable
	if err != nil {
		return "", fmt.Errorf("fetching deploy data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching deploy data for owner %q: unexpected status %s", owner, resp.Status)
	}

	var deployData struct {
		TenantName string `json:"TENANT_NAME"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deployData); err != nil {
		return "", fmt.Errorf("decoding deploy data: %w", err)
	}

	if deployData.TenantName == "" {
		return "", fmt.Errorf("cannot discover tenant name automatically for %q - set NAIS_API_TENANT", owner)
	}

	return deployData.TenantName, nil
}

func githubTokenSource(ctx context.Context, requestURL, requestToken string) oauth2.TokenSource {
	return tokenSourceFunc(func() (*oauth2.Token, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil) // #nosec G704 -- requestURL comes from ACTIONS_ID_TOKEN_REQUEST_URL env var set by GitHub Actions; SSRF risk is acceptable
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		q := req.URL.Query()
		q.Add("audience", "api.nais.io")

		req.URL.RawQuery = q.Encode()
		req.Header.Set("Authorization", "bearer "+requestToken)

		resp, err := http.DefaultClient.Do(req) // #nosec G704 -- request URL originates from ACTIONS_ID_TOKEN_REQUEST_URL; SSRF risk is acceptable
		if err != nil {
			return nil, fmt.Errorf("fetching token: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

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

		// Skip signature verification here as we're only interested in the expiry time.
		// The API will validate the token later.
		j, err := jwt.ParseString(
			tokenResponse.Token,
			jwt.WithVerify(false),
			jwt.WithAcceptableSkew(10*time.Second),
		)
		if err != nil {
			return nil, fmt.Errorf("parse jwt: %w", err)
		}

		expiry, ok := j.Expiration()
		if !ok {
			return nil, fmt.Errorf("missing expiry claim")
		}

		return &oauth2.Token{
			AccessToken: tokenResponse.Token,
			Expiry:      expiry,
		}, nil
	})
}
