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

// EnsureSecretAccess creates an elevation to allow reading the database secret.
// This is required because users don't have direct access to secrets without elevation.
// The elevation is logged for audit purposes.
func EnsureSecretAccess(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, reason string, out *naistrix.OutputWriter) error {
	if reason == "" {
		return fmt.Errorf("reason is required for accessing database secrets")
	}

	// Load kubeconfig to get defaults for namespace and context if not provided
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: string(cluster),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Determine team slug from namespace
	team := string(namespace)
	if team == "" {
		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("unable to get namespace from kubeconfig: %w", err)
		}
		team = ns
	}
	if team == "" {
		return fmt.Errorf("namespace is required to determine team (use --namespace flag or set in kubeconfig)")
	}

	// Determine environment from kubeconfig context
	environmentName := string(cluster)
	if environmentName == "" {
		rawConfig, err := kubeConfig.RawConfig()
		if err != nil {
			return fmt.Errorf("unable to get kubeconfig: %w", err)
		}
		environmentName = rawConfig.CurrentContext
	}
	if environmentName == "" {
		return fmt.Errorf("kubeconfig context is required to determine environment (use --context flag or set current-context in kubeconfig)")
	}

	// The secret name follows the pattern "google-sql-<appname>"
	secretName := "google-sql-" + appName

	out.Debugf("Requesting elevated access to secret %q for database connection...\n", secretName)

	_, err := naisapi.CreateElevation(ctx, team, environmentName, secretName, reason, 5)
	if err != nil {
		// Check if the error indicates the user is not authorized
		if strings.Contains(err.Error(), "not authorized") || strings.Contains(err.Error(), "Not authorized") {
			return fmt.Errorf("you are not authorized to access this database. Make sure you are a member of team %q", team)
		}
		return fmt.Errorf("creating elevation for secret access: %w", err)
	}

	out.Debugf("✅ Access granted.\n")
	return nil
}

// EnsureSecretAccessWithUserReason creates an elevation with a user-provided reason.
// This should be used for interactive operations like proxy and psql where the user
// should provide justification for accessing the database.
func EnsureSecretAccessWithUserReason(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, reason string, out *naistrix.OutputWriter) error {
	if reason == "" {
		return fmt.Errorf("reason is required for accessing database secrets (use --reason flag)")
	}

	if len(reason) < 10 {
		return fmt.Errorf("reason must be at least 10 characters")
	}

	return EnsureSecretAccess(ctx, appName, namespace, cluster, reason, out)
}
