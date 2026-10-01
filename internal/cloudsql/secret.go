package cloudsql

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

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

type SecretValues struct{ values map[string]string }

func (s *SecretValues) Get(suffix string) string {
	for name, val := range s.values {
		if strings.HasSuffix(name, suffix) {
			return val
		}
	}
	return ""
}

// GetSecretValues retrieves Cloud SQL secret values through the audited API.
func GetSecretValues(ctx context.Context, appName, team, environment string, fl *flag.CloudSQL, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	if reason == "" {
		reason = fl.Reason
		if reason == "" {
			return nil, fmt.Errorf("reason is required for accessing database secrets")
		}
	}
	out.Printf("Using team %q\n", team)
	secretName := "google-sql-" + appName
	out.Debugf("Requesting access to Cloud SQL secret %q...\n", secretName)
	values, err := naisapi.ViewSecretValues(ctx, team, environment, secretName, reason)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") || strings.Contains(err.Error(), "Not authorized") {
			return nil, fmt.Errorf("you are not authorized to access this database. Make sure you are a member of team %q", team)
		}
		return nil, err
	}
	out.Debugf("✅ Access granted.\n")
	result := &SecretValues{values: make(map[string]string, len(values))}
	for _, v := range values {
		result.values[v.Name] = v.Value
	}
	return result, nil
}

func GetSecretValuesWithUserReason(ctx context.Context, appName, team, environment string, fl *flag.CloudSQL, reason string, out *naistrix.OutputWriter) (*SecretValues, error) {
	if reason == "" {
		reason = fl.Reason
		if reason == "" {
			return nil, fmt.Errorf("reason is required for accessing database secrets (use --reason flag)")
		}
	}
	if len(reason) < 10 {
		return nil, fmt.Errorf("reason must be at least 10 characters")
	}
	return GetSecretValues(ctx, appName, team, environment, fl, reason, out)
}
