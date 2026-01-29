package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"k8s.io/client-go/tools/clientcmd"
)

// Hardcoded reasons for administrative operations
const (
	ReasonPasswordRotate = "Rotating database password via nais CLI"
	ReasonPrepareAccess  = "Preparing database for IAM user access via nais CLI"
	ReasonRevokeAccess   = "Revoking IAM user access from database via nais CLI"
	ReasonListUsers      = "Listing database users via nais CLI"
	ReasonAddUser        = "Adding database user via nais CLI"
	ReasonDropUser       = "Dropping database user via nais CLI"
	ReasonEnableAudit    = "Enabling audit logging via nais CLI"
	ReasonVerifyAudit    = "Verifying audit configuration via nais CLI"
)

// SecretValues holds the secret values retrieved from the API
type SecretValues struct {
	values map[string]string
}

// Get returns the value for a given key, or empty string if not found
func (s *SecretValues) Get(key string) string {
	if s == nil || s.values == nil {
		return ""
	}
	return s.values[key]
}

// GetBySuffix returns the value for a key that ends with the given suffix
func (s *SecretValues) GetBySuffix(suffix string) string {
	if s == nil || s.values == nil {
		return ""
	}
	for k, v := range s.values {
		if strings.HasSuffix(k, suffix) {
			return v
		}
	}
	return ""
}

// resolveTeamAndEnvironment extracts team and environment from namespace and cluster flags
func resolveTeamAndEnvironment(namespace flag.Namespace, cluster flag.Context) (team, environment string, err error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: string(cluster),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Determine team slug from namespace
	team = string(namespace)
	if team == "" {
		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return "", "", fmt.Errorf("unable to get namespace from kubeconfig: %w", err)
		}
		team = ns
	}
	if team == "" {
		return "", "", fmt.Errorf("namespace is required to determine team (use --namespace flag or set in kubeconfig)")
	}

	// Determine environment from kubeconfig context
	environment = string(cluster)
	if environment == "" {
		rawConfig, err := kubeConfig.RawConfig()
		if err != nil {
			return "", "", fmt.Errorf("unable to get kubeconfig: %w", err)
		}
		environment = rawConfig.CurrentContext
	}
	if environment == "" {
		return "", "", fmt.Errorf("kubeconfig context is required to determine environment (use --context flag or set current-context in kubeconfig)")
	}

	return team, environment, nil
}

// GetSecretValues retrieves the values of a database secret via the API.
// This is the preferred method for accessing secret values as it combines
// authorization, logging, and value retrieval in a single operation.
// The access is logged for audit purposes.
func GetSecretValues(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	if reason == "" {
		return nil, fmt.Errorf("reason is required for accessing database secrets")
	}

	team, environmentName, err := resolveTeamAndEnvironment(namespace, cluster)
	if err != nil {
		return nil, err
	}

	// The secret name follows the pattern "google-sql-<appname>"
	secretName := "google-sql-" + appName

	out.Debugf("Requesting access to secret %q for database connection...\n", secretName)

	values, err := naisapi.ViewSecretValues(ctx, team, environmentName, secretName, reason)
	if err != nil {
		// Check if the error indicates the user is not authorized
		if strings.Contains(err.Error(), "not authorized") || strings.Contains(err.Error(), "Not authorized") {
			return nil, fmt.Errorf("you are not authorized to access this database. Make sure you are a member of team %q", team)
		}
		return nil, fmt.Errorf("viewing secret values: %w", err)
	}

	out.Debugf("✅ Access granted.\n")

	// Convert to SecretValues
	result := &SecretValues{
		values: make(map[string]string, len(values)),
	}
	for _, v := range values {
		result.values[v.Name] = v.Value
	}

	return result, nil
}

// GetSecretValuesWithUserReason retrieves secret values with a user-provided reason.
// This should be used for interactive operations like proxy and psql where the user
// should provide justification for accessing the database.
func GetSecretValuesWithUserReason(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	if reason == "" {
		return nil, fmt.Errorf("reason is required for accessing database secrets (use --reason flag)")
	}

	if len(reason) < 10 {
		return nil, fmt.Errorf("reason must be at least 10 characters")
	}

	return GetSecretValues(ctx, appName, namespace, cluster, reason, out)
}
