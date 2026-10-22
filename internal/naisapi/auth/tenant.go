package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

var tenantNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// resolveConsoleHost preserves automatic tenant discovery for tenant users.
// Only nais.io users use NAIS_API_TENANT, evaluated on every command so they can
// switch tenants without replacing their stored login session.
func resolveConsoleHost(ctx context.Context, domain, storedHost string) (string, error) {
	if domain == "nais.io" {
		tenant := os.Getenv("NAIS_API_TENANT")
		if tenant == "" {
			return "", fmt.Errorf("set NAIS_API_TENANT to select a tenant when using a nais.io account (for example, NAIS_API_TENANT=nav)")
		}
		if !tenantNamePattern.MatchString(tenant) {
			return "", fmt.Errorf("invalid NAIS_API_TENANT %q: expected a tenant name such as nav or dev-nais", tenant)
		}
		return "console." + tenant + ".cloud.nais.io", nil
	}
	if storedHost != "" {
		return storedHost, nil
	}

	u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil) // #nosec G704 -- Fixed GCS base URL; domain originates from the verified ID token.
	if err != nil {
		return "", fmt.Errorf("create tenant data request: %w", err)
	}
	res, err := http.DefaultClient.Do(req) // #nosec G704 -- Fixed GCS base URL; domain originates from the verified ID token.
	if err != nil {
		return "", fmt.Errorf("get tenant data: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no tenant data found for domain %q", domain)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get tenant data: unexpected status code %s", res.Status)
	}
	var tenant struct {
		ConsoleURL string `json:"consoleUrl"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tenant); err != nil {
		return "", fmt.Errorf("decode tenant data: %w", err)
	}
	if tenant.ConsoleURL == "" {
		return "", fmt.Errorf("tenant data for domain %q is missing consoleUrl", domain)
	}
	return tenant.ConsoleURL, nil
}
