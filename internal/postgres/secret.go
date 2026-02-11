package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

// Default duration for in-cluster postgres access grants
const defaultPostgresAccessDuration = "1h"

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

// isCloudSQLDatabase checks if the given app uses CloudSQL or in-cluster postgres
func isCloudSQLDatabase(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context) (bool, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: string(cluster),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	ns := string(namespace)
	if ns == "" {
		var err error
		ns, _, err = kubeConfig.Namespace()
		if err != nil {
			return false, fmt.Errorf("unable to get namespace from kubeconfig: %w", err)
		}
	}

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return false, fmt.Errorf("unable to get kubeconfig: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return false, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	// Check for CloudSQL SQLInstance resources
	sqlInstances, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(ns).List(ctx, v1.ListOptions{
		LabelSelector: "app=" + appName,
	})
	if err != nil {
		return false, fmt.Errorf("error looking for sqlinstance %q in %q: %w", appName, ns, err)
	}

	return len(sqlInstances.Items) >= 1, nil
}

// getPostgresClusterName retrieves the postgres cluster name for an app
func getPostgresClusterName(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context) (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: string(cluster),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	ns := string(namespace)
	if ns == "" {
		var err error
		ns, _, err = kubeConfig.Namespace()
		if err != nil {
			return "", fmt.Errorf("unable to get namespace from kubeconfig: %w", err)
		}
	}

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return "", fmt.Errorf("unable to get kubeconfig: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("unable to create dynamic client: %w", err)
	}

	// First try to get the cluster name from the Application spec
	unstructuredApp, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "nais.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).Namespace(ns).Get(ctx, appName, v1.GetOptions{})
	if err == nil {
		spec, ok := unstructuredApp.Object["spec"].(map[string]interface{})
		if ok {
			postgres, ok := spec["postgres"].(map[string]interface{})
			if ok {
				clusterName, ok := postgres["clusterName"].(string)
				if ok && clusterName != "" {
					return clusterName, nil
				}
			}
		}
	}

	// If no Application found or no clusterName in spec, check if there's a Postgres resource with this name
	_, err = dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "data.nais.io",
		Version:  "v1",
		Resource: "postgres",
	}).Namespace(ns).Get(ctx, appName, v1.GetOptions{})
	if err == nil {
		// The appName is actually a postgres cluster name
		return appName, nil
	}

	return "", fmt.Errorf("unable to find postgres cluster for %q in %q", appName, ns)
}

// GetSecretValues retrieves the values of a database secret via the API.
// For CloudSQL databases, this retrieves the secret values directly.
// For in-cluster postgres, this grants temporary access to the database.
// In both cases, the access is logged for audit purposes.
func GetSecretValues(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	if reason == "" {
		return nil, fmt.Errorf("reason is required for accessing database secrets")
	}

	team, environmentName, err := resolveTeamAndEnvironment(namespace, cluster)
	if err != nil {
		return nil, err
	}

	// Check if this is a CloudSQL or in-cluster postgres database
	isCloudSQL, err := isCloudSQLDatabase(ctx, appName, namespace, cluster)
	if err != nil {
		return nil, fmt.Errorf("checking database type: %w", err)
	}

	if isCloudSQL {
		return getCloudSQLSecretValues(ctx, appName, team, environmentName, reason, out)
	}

	return grantInClusterPostgresAccess(ctx, appName, namespace, cluster, team, environmentName, reason, out)
}

// getCloudSQLSecretValues retrieves secret values for CloudSQL databases
func getCloudSQLSecretValues(ctx context.Context, appName, team, environmentName, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	// The secret name follows the pattern "google-sql-<appname>"
	secretName := "google-sql-" + appName

	out.Debugf("Requesting access to CloudSQL secret %q...\n", secretName)

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

// grantPostgresAccess grants temporary access to an in-cluster postgres database.
// This creates a time-limited grant for the user and logs the access for auditing purposes.
func grantPostgresAccess(ctx context.Context, clusterName, teamSlug, environmentName, grantee, duration string) error {
	_ = `# @genqlient
mutation GrantPostgresAccess($input: GrantPostgresAccessInput!) {
	grantPostgresAccess(input: $input) {
		error
	}
}
`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return fmt.Errorf("creating GraphQL client: %w", err)
	}

	resp, err := gql.GrantPostgresAccess(ctx, client, gql.GrantPostgresAccessInput{
		ClusterName:     clusterName,
		TeamSlug:        teamSlug,
		EnvironmentName: environmentName,
		Grantee:         grantee,
		Duration:        duration,
	})
	if err != nil {
		return fmt.Errorf("granting postgres access: %w", err)
	}

	if resp.GrantPostgresAccess.Error != "" {
		return fmt.Errorf("granting postgres access: %s", resp.GrantPostgresAccess.Error)
	}

	return nil
}

// grantInClusterPostgresAccess grants access to in-cluster postgres databases
func grantInClusterPostgresAccess(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, team, environmentName, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	// Get the postgres cluster name
	clusterName, err := getPostgresClusterName(ctx, appName, namespace, cluster)
	if err != nil {
		return nil, err
	}

	// Get the authenticated user's email
	user, err := naisapi.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting authenticated user: %w", err)
	}
	grantee := user.Email()

	out.Debugf("Requesting access to in-cluster postgres %q for user %q...\n", clusterName, grantee)

	// Grant access via the API (this logs the access for audit purposes)
	err = grantPostgresAccess(ctx, clusterName, team, environmentName, grantee, defaultPostgresAccessDuration)
	if err != nil {
		// Check if the error indicates the user is not authorized
		if strings.Contains(err.Error(), "not authorized") || strings.Contains(err.Error(), "Not authorized") {
			return nil, fmt.Errorf("you are not authorized to access this database. Make sure you are a member of team %q", team)
		}
		return nil, fmt.Errorf("granting postgres access: %w", err)
	}

	out.Debugf("✅ Access granted for %s.\n", defaultPostgresAccessDuration)

	// For in-cluster postgres, we don't return secret values as authentication
	// happens via OAuth tokens, not via secrets
	return &SecretValues{values: make(map[string]string)}, nil
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
