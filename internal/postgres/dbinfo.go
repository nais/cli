package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nais/cli/internal/k8s"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func NewDBInfo(appName, namespace string, context k8s.Context) (*DBInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: string(context),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: unable to get kubeconfig: %w", err)
	}

	if namespace == "" {
		namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("NewDBInfo: unable to get namespace: %w", err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: load kubeclient configuration: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: load kubeclient configuration: %w", err)
	}

	return &DBInfo{
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
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

	return createConnectionInfo(*secret, connectionName), nil
}

func createConnectionInfo(secret corev1.Secret, instance string) *ConnectionInfo {
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
		host:     getSecretDataValue(secret, "_HOST"),
		url:      pgUrl,
		jdbcUrl:  jdbcUrl,
		instance: instance,
	}
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

	connectionName, ok, err := unstructured.NestedString(sqlInstance.Object, "status", "connectionName")
	if !ok || err != nil {
		return fmt.Errorf("missing 'connectionName' status field; run 'kubectl describe sqlinstance %s' and check for status failures", sqlInstance.GetName())
	}

	i.connectionName = connectionName
	i.projectID = sqlInstance.GetAnnotations()["cnrm.cloud.google.com/project-id"]
	return nil
}

type ConnectionInfo struct {
	username string
	password string
	dbName   string
	instance string
	port     string
	host     string
	url      *url.URL
	jdbcUrl  *url.URL
}

func (c *ConnectionInfo) ProxyConnectionString() string {
	return fmt.Sprintf("host=%v user=%v dbname=%v password=%v sslmode=disable", c.instance, c.username, c.dbName, c.password)
}

func (c *ConnectionInfo) SetPassword(password string) {
	c.password = password
	if c.url != nil {
		c.url.User = url.UserPassword(c.username, password)
	}
	if c.jdbcUrl != nil {
		queries := c.jdbcUrl.Query()
		queries.Set("password", password)
		c.jdbcUrl.RawQuery = queries.Encode()
	} else if c.url != nil {
		queries := c.url.Query()
		queries.Set("password", password)
		queries.Set("user", c.username)
		c.jdbcUrl = &url.URL{
			Scheme:   "jdbc:postgresql",
			Host:     c.url.Host,
			Path:     c.dbName,
			RawQuery: queries.Encode(),
		}
	}
}

func getSecretDataValue(secret corev1.Secret, suffix string) string {
	for name, val := range secret.Data {
		if strings.HasSuffix(name, suffix) {
			return string(val)
		}
	}
	return ""
}

// formatInvalidGrantError returns a custom error message if the error is of type oauth2.RetrieveError and if it has the
// error code invalid_grant. If not it returns the error.
func formatInvalidGrantError(err error) error {
	var retrieve *oauth2.RetrieveError
	if errors.As(err, &retrieve) {
		if retrieve.ErrorCode == "invalid_grant" {
			return fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth login --update-adc` first")
		}
	}

	return err
}
