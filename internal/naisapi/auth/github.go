package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func GithubActions(ctx context.Context) (*AuthenticatedUser, bool, error) {
	u := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	tok := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if u == "" || tok == "" {
		return nil, false, nil
	}

	// TODO: this is temporary; should be derived from the token somehow
	tenant := os.Getenv("NAIS_API_TENANT")
	if tenant == "" {
		return nil, false, fmt.Errorf("NAIS_API_TENANT must be set when using GitHub Actions authentication")
	}

	ts := &githubActionsTokenSource{ctx, u, tok}
	token, err := ts.Token()
	if err != nil {
		return nil, false, fmt.Errorf("getting token from github actions: %w", err)
	}

	// TODO: this should be a oauth2.ReuseTokenSource that refreshes the token when expired
	return &AuthenticatedUser{
		TokenSource: oauth2.StaticTokenSource(token),
		consoleHost: "console." + tenant + ".cloud.nais.io",
		domain:      "TODO",
		email:       "TODO",
	}, true, nil
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
