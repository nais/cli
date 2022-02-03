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
		return nil, err
	}

	if namespace == "" {
		namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			return nil, err
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
		return nil, err
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
	app, err := i.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(i.namespace).Get(ctx, i.appName, v1.GetOptions{})
	if err != nil {
		return err
	}

	i.connectionName = app.Object["status"].(map[string]interface{})["connectionName"].(string)
	i.projectID = app.GetAnnotations()["cnrm.cloud.google.com/project-id"]
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
