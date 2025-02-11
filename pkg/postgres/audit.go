package postgres

import (
	"context"
	"database/sql"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func EnableAuditLogging(ctx context.Context, appName, cluster, namespace string) error {
	return enableAuditAsAppUser(ctx, appName, namespace, cluster)
}

func enableAuditAsAppUser(ctx context.Context, appName, namespace, cluster string) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	err = validateAuditFlags(ctx, dbInfo)
	if err != nil {
		return fmt.Errorf("required flags missing for instance: %v", err)
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}

	defer db.Close()

	enableAudit := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS pgaudit; ALTER USER %s IN DATABASE %s SET pgaudit.log TO 'none';`, connectionInfo.username, connectionInfo.dbName)
	_, err = db.ExecContext(ctx, enableAudit)
	if err != nil {
		return fmt.Errorf("enableAuditAsAppUser: error enabling pgaudit: %w", err)
	}

	return nil
}

func validateAuditFlags(ctx context.Context, info *DBInfo) error {
	dbFlags, err := getDBFlags(ctx, info)
	if err != nil {
		return fmt.Errorf("validateAuditFlags: error getting db flags: %w", err)
	}

	requiredFlags := []string{
		"cloudsql.enable_pgaudit",
		"pgaudit.log",
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

func getDBFlags(ctx context.Context, info *DBInfo) (map[string]string, error) {
	dbFlags := make(map[string]string)
	sqlInstances, err := info.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(info.namespace).List(ctx, v1.ListOptions{
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
