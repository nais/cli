package postgres

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DBInfo struct {
	k8sClient      kubernetes.Interface
	dynamicClient  dynamic.Interface
	config         clientcmd.ClientConfig
	namespace      string
	appName        string
	projectID      string
	connectionName string
}

func NewDBInfo(appName, namespace, context string) (*DBInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: context,
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: unable to get kubeconfig: %w", err)
	}

	if namespace == "" {
		namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("NewDBConfig: unable to get namespace: %w", err)
		}
	}

	return &DBInfo{
		k8sClient:     kubernetes.NewForConfigOrDie(config),
		dynamicClient: dynamic.NewForConfigOrDie(config),
		config:        kubeConfig,
		namespace:     namespace,
		appName:       appName,
	}, nil
}

func (i *DBInfo) ProjectID(ctx context.Context) (string, error) {
	if i.projectID == "" {
		err := i.fetchDBInstance(ctx)
		if err != nil {
			return "", err
		}
	}
	return i.projectID, nil
}

func (i *DBInfo) ConnectionName(ctx context.Context) (string, error) {
	if i.connectionName == "" {
		err := i.fetchDBInstance(ctx)
		if err != nil {
			return "", err
		}
	}
	return i.connectionName, nil
}

func (i *DBInfo) DBConnection(ctx context.Context) (*ConnectionInfo, error) {
	secret, err := i.k8sClient.CoreV1().Secrets(i.namespace).Get(ctx, "google-sql-"+i.appName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get database password from %q in %q: %w", "google-sql-"+i.appName, i.namespace, err)
	}

	connectionName, err := i.ConnectionName(ctx)
	if err != nil {
		return nil, err
	}

	return &ConnectionInfo{
		username: getSecretDataValue(secret, "_USERNAME"),
		password: getSecretDataValue(secret, "_PASSWORD"),
		dbName:   getSecretDataValue(secret, "_DATABASE"),
		host:     connectionName,
	}, nil
}

func (i *DBInfo) fetchDBInstance(ctx context.Context) error {
	sqlInstances, err := i.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(i.namespace).List(ctx, v1.ListOptions{
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

	connectionName, ok := sqlInstance.Object["status"].(map[string]interface{})["connectionName"]
	if !ok {
		return fmt.Errorf("missing 'connectionName' status field; run 'kubectl describe sqlinstance %s' and check for status failures", sqlInstance.GetName())
	}

	i.connectionName = connectionName.(string)
	i.projectID = sqlInstance.GetAnnotations()["cnrm.cloud.google.com/project-id"]
	return nil
}

type ConnectionInfo struct {
	username string
	password string
	dbName   string
	host     string
}

func (c *ConnectionInfo) ConnectionString() string {
	return fmt.Sprintf("host=%v user=%v dbname=%v password=%v sslmode=disable", c.host, c.username, c.dbName, c.password)
}
