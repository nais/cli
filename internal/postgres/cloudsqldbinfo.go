package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	v2 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CloudSQLDBInfo struct {
	DBInfo
	projectID      string
	connectionName string
}

func (i *CloudSQLDBInfo) ToCloudSQLDBInfo() (*CloudSQLDBInfo, error) {
	return i, nil
}

func (i *CloudSQLDBInfo) ProjectID(ctx context.Context) (string, error) {
	if i.projectID == "" {
		err := i.fetchDBInstance(ctx)
		if err != nil {
			return "", err
		}
	}
	return i.projectID, nil
}

func (i *CloudSQLDBInfo) ConnectionName(ctx context.Context) (string, error) {
	if i.connectionName == "" {
		err := i.fetchDBInstance(ctx)
		if err != nil {
			return "", err
		}
	}
	return i.connectionName, nil
}

func (i *CloudSQLDBInfo) DBConnection(ctx context.Context) (*ConnectionInfo, error) {
	secret, err := i.k8sClient.CoreV1().Secrets(string(i.namespace)).Get(ctx, "google-sql-"+i.appName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get database password from %q in %q: %w", "google-sql-"+i.appName, i.namespace, err)
	}

	connectionName, err := i.ConnectionName(ctx)
	if err != nil {
		return nil, err
	}

	return createConnectionInfo(*secret, connectionName), nil
}

func createConnectionInfo(secret v2.Secret, instance string) *ConnectionInfo {
	var pgUrl *url.URL
	var jdbcUrl *url.URL
	var err error
	for name, val := range secret.Data {
		if strings.HasSuffix(name, "_URL") {
			value := string(val)
			if strings.HasSuffix(name, "_JDBC_URL") {
				jdbcUrl, err = url.Parse(value)
			} else {
				pgUrl, err = url.Parse(value)
			}
			if err != nil {
				panic(err)
			}
		}
	}

	return &ConnectionInfo{
		username: getSecretDataValue(secret, "_USERNAME"),
		password: getSecretDataValue(secret, "_PASSWORD"),
		dbName:   getSecretDataValue(secret, "_DATABASE"),
		port:     getSecretDataValue(secret, "_PORT"),
		url:      pgUrl,
		jdbcUrl:  jdbcUrl,
		instance: instance,
	}
}

func getSecretDataValue(secret v2.Secret, suffix string) string {
	for name, val := range secret.Data {
		if strings.HasSuffix(name, suffix) {
			return string(val)
		}
	}
	return ""
}

func (i *CloudSQLDBInfo) fetchDBInstance(ctx context.Context) error {
	sqlInstances, err := i.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(string(i.namespace)).List(ctx, v1.ListOptions{
		LabelSelector: "app=" + i.appName,
	})
	if err != nil {
		return fmt.Errorf("fetchDBInstance: error looking for sqlinstance %q in %q: %w", i.appName, i.namespace, err)
	}

	if len(sqlInstances.Items) == 0 {
		return fmt.Errorf("fetchDBInstance: no sqlinstance found for app %q in %q", i.appName, i.namespace)
	} else if len(sqlInstances.Items) > 1 {
		return fmt.Errorf("fetchDBInstance: multiple sqlinstances found for app %q in %q", i.appName, i.namespace)
	}

	sqlInstance := sqlInstances.Items[0]

	connectionName, ok, err := unstructured.NestedString(sqlInstance.Object, "status", "connectionName")
	if !ok || err != nil {
		return fmt.Errorf("missing 'connectionName' status field; run 'kubectl describe sqlinstance %s' and check for status failures", sqlInstance.GetName())
	}

	i.connectionName = connectionName
	i.projectID = sqlInstance.GetAnnotations()["cnrm.cloud.google.com/project-id"]
	return nil
}
