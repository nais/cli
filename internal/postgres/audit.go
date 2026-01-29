package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func EnableAuditLogging(ctx context.Context, appName string, cluster flag.Context, namespace flag.Namespace, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	if _, err := GetSecretValues(ctx, appName, namespace, cluster, ReasonEnableAudit, out); err != nil {
		return err
	}
	return enableAuditAsAppUser(ctx, appName, namespace, cluster, out)
}

func VerifyAuditLogging(ctx context.Context, appName string, cluster flag.Context, namespace flag.Namespace, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	if _, err := GetSecretValues(ctx, appName, namespace, cluster, ReasonVerifyAudit, out); err != nil {
		return err
	}
	_, err := verifyAuditAsAppUser(ctx, appName, namespace, cluster, out)
	return err
}

func enableAuditAsAppUser(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, out *naistrix.OutputWriter) error {
	dbInfo, err := NewDBInfo(ctx, appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	cloudSQLDbInfo, err := dbInfo.ToCloudSQLDBInfo()
	if err != nil {
		return err
	}

	err = validateAuditFlags(ctx, cloudSQLDbInfo)
	if err != nil {
		return fmt.Errorf("required flags missing for instance: %v", err)
	}

	isConfigured, err := checkAuditConfigured(ctx, connectionInfo)
	if err != nil {
		return fmt.Errorf("error checking audit configuration: %w", err)
	}

	if isConfigured {
		out.Println("✅ Audit is already properly configured. No changes needed.")
		return nil
	}

	// If we get here, we need to enable audit
	out.Println("Audit configuration needs to be updated...\n")

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pgaudit")
	if err != nil {
		return fmt.Errorf("enableAuditAsAppUser: error creating pgaudit extension: %w", err)
	}

	alterUserQuery := fmt.Sprintf(
		"ALTER USER %s IN DATABASE %s SET pgaudit.log TO 'none'",
		pq.QuoteIdentifier(connectionInfo.username),
		pq.QuoteIdentifier(connectionInfo.dbName),
	)
	_, err = db.ExecContext(ctx, alterUserQuery)
	if err != nil {
		return fmt.Errorf("enableAuditAsAppUser: error configuring pgaudit.log: %w", err)
	}

	out.Println("✅ Successfully enabled audit extension and configured pgaudit.log for application user")
	return nil
}

func checkAuditConfigured(ctx context.Context, connectionInfo *ConnectionInfo) (bool, error) {
	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return false, err
	}
	defer db.Close()

	// Check if pgaudit extension is installed
	var extensionExists bool
	checkExtensionQuery := "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pgaudit')"
	err = db.QueryRowContext(ctx, checkExtensionQuery).Scan(&extensionExists)
	if err != nil {
		return false, err
	}

	if !extensionExists {
		return false, nil
	}

	// Check pgaudit.log setting for the application user
	var pgauditLogValue string
	checkSettingQuery := "SELECT setting FROM pg_settings WHERE name = 'pgaudit.log'"
	err = db.QueryRowContext(ctx, checkSettingQuery).Scan(&pgauditLogValue)
	if err != nil {
		return false, err
	}

	// Check if it's set to 'none'
	return pgauditLogValue == "none", nil
}

func validateAuditFlags(ctx context.Context, info *CloudSQLDBInfo) error {
	dbFlags, err := getDBFlags(ctx, info)
	if err != nil {
		return fmt.Errorf("validateAuditFlags: error getting db flags: %w", err)
	}

	requiredFlags := []string{
		"cloudsql.enable_pgaudit",
		"pgaudit.log",
		"pgaudit.log_parameter",
	}

	err = validateRequiredFlags(dbFlags, requiredFlags)
	if err != nil {
		return fmt.Errorf("validateAuditFlags: %v", err)
	}
	return nil
}

func validateRequiredFlags(dbFlags map[string]string, requiredFlags []string) error {
	for _, reqFlag := range requiredFlags {
		if _, exists := dbFlags[reqFlag]; !exists {
			return fmt.Errorf("required flag %q missing", reqFlag)
		}
	}

	return nil
}

func getDBFlags(ctx context.Context, info *CloudSQLDBInfo) (map[string]string, error) {
	dbFlags := make(map[string]string)
	sqlInstances, err := info.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(string(info.namespace)).List(ctx, v1.ListOptions{
		LabelSelector: "app=" + info.appName,
	})
	if err != nil {
		return dbFlags, fmt.Errorf("GetDBInstance: error looking for sqlinstance %q in %q: %w", info.appName, info.namespace, err)
	}

	if len(sqlInstances.Items) != 1 {
		return dbFlags, fmt.Errorf("GetDBInstance: expected one sqlinstance for app %q in %q, got %d", info.appName, info.namespace, len(sqlInstances.Items))
	}

	spec, ok := sqlInstances.Items[0].Object["spec"].(map[string]interface{})
	if !ok {
		return dbFlags, fmt.Errorf("GetDBInstance: error accessing spec for app %q in %q", info.appName, info.namespace)
	}

	settings, ok := spec["settings"].(map[string]interface{})
	if !ok {
		return dbFlags, fmt.Errorf("GetDBInstance: error accessing settings for app %q in %q", info.appName, info.namespace)
	}

	databaseFlags, ok := settings["databaseFlags"].([]interface{})
	if !ok {
		return dbFlags, fmt.Errorf("GetDBInstance: error accessing databaseFlags for app %q in %q", info.appName, info.namespace)
	}

	for _, flag := range databaseFlags {
		f, ok := flag.(map[string]interface{})
		if !ok {
			return dbFlags, fmt.Errorf("GetDBInstance: error accessing databaseFlags for app %q in %q", info.appName, info.namespace)
		}
		name, nameOk := f["name"].(string)
		value, valueOk := f["value"].(string)
		if nameOk && valueOk {
			dbFlags[name] = value
		}

	}

	return dbFlags, nil
}

func verifyAuditAsAppUser(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, out *naistrix.OutputWriter) (bool, error) {
	dbInfo, err := NewDBInfo(ctx, appName, namespace, cluster)
	if err != nil {
		return false, err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return false, err
	}

	cloudSQLDbInfo, err := dbInfo.ToCloudSQLDBInfo()
	if err != nil {
		return false, err
	}

	out.Println("\nVerifying audit configuration for application: <info>" + appName + "</info>\n")

	dbFlags, err := getDBFlags(ctx, cloudSQLDbInfo)
	if err != nil {
		return false, fmt.Errorf("error getting db flags: %w", err)
	}

	enablePgaudit, enableExists := dbFlags["cloudsql.enable_pgaudit"]
	if !enableExists {
		out.Println("  ❌ Flag <info>cloudsql.enable_pgaudit</info> is missing")
		return false, fmt.Errorf("cloudsql.enable_pgaudit flag is not set")
	}
	if enablePgaudit != "on" && enablePgaudit != "true" {
		out.Printf("  ❌ Flag <info>cloudsql.enable_pgaudit</info>: expected <info>on</info> or <info>true</info>, got <info>%s</info>\n", enablePgaudit)
		return false, fmt.Errorf("cloudsql.enable_pgaudit must be set to 'on' or 'true'")
	}
	out.Printf("  ✅ Flag <info>cloudsql.enable_pgaudit</info> = <info>%s</info>\n", enablePgaudit)

	pgauditLog, logExists := dbFlags["pgaudit.log"]
	if !logExists {
		out.Println("  ❌ Flag <info>pgaudit.log</info> is missing")
		return false, fmt.Errorf("pgaudit.log flag is not set")
	}
	out.Printf("  ✅ Flag <info>pgaudit.log</info> = <info>%s</info>\n", pgauditLog)

	logParameter, paramExists := dbFlags["pgaudit.log_parameter"]
	if !paramExists {
		out.Println("  ❌ Flag <info>pgaudit.log_parameter</info> is missing")
		return false, fmt.Errorf("pgaudit.log_parameter flag is not set")
	}
	if logParameter != "on" && logParameter != "true" {
		out.Printf("  ❌ Flag <info>pgaudit.log_parameter</info>: expected <info>on</info> or <info>true</info>, got <info>%s</info>\n", logParameter)
		return false, fmt.Errorf("pgaudit.log_parameter must be set to 'on' or 'true'")
	}
	out.Printf("  ✅ Flag <info>pgaudit.log_parameter</info> = <info>%s</info>\n", logParameter)

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return false, fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		return false, fmt.Errorf("error pinging database: %w", err)
	}

	var extensionExists bool
	checkExtensionQuery := "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pgaudit')"
	err = db.QueryRowContext(ctx, checkExtensionQuery).Scan(&extensionExists)
	if err != nil {
		return false, fmt.Errorf("error checking pgaudit extension: %w", err)
	}

	if !extensionExists {
		out.Println("\n  ❌ pgaudit extension is not installed")
		return false, nil
	}
	out.Println("\n  ✅ pgaudit extension is installed")

	var pgauditLogValue string
	checkSettingQuery := "SELECT setting FROM pg_settings WHERE name = 'pgaudit.log'"
	err = db.QueryRowContext(ctx, checkSettingQuery).Scan(&pgauditLogValue)
	if err != nil {
		return false, fmt.Errorf("error checking pgaudit.log setting from pg_settings: %w", err)
	}

	expectedValue := "none"
	if pgauditLogValue != expectedValue {
		out.Printf("  ❌ pgaudit.log setting for application user: expected <info>%s</info>, got <info>%s</info>\n", expectedValue, pgauditLogValue)
		return false, nil
	}

	out.Printf("  ✅ pgaudit.log setting for application user: <info>%s</info>\n", pgauditLogValue)
	out.Println("\n✅ All audit configurations are correct!")

	return true, nil
}
