package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"golang.org/x/oauth2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DB interface {
	DBConnection(ctx context.Context) (*ConnectionInfo, error)
	RunProxy(ctx context.Context, host string, port *uint, portCh chan<- int, out *naistrix.OutputWriter, printInstructions bool) error

	AppName() string

	// TODO: Remove when interface migration complete
	ToCloudSQLDBInfo() (*CloudSQLDBInfo, error)
}

type DBInfo struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        clientcmd.ClientConfig
	namespace     flag.Namespace
	appName       string
}

func (d *DBInfo) AppName() string {
	return d.appName
}

func NewDBInfo(ctx context.Context, appName string, namespace flag.Namespace, context flag.Context) (DB, error) {
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
		ns, _, err := kubeConfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("NewDBInfo: unable to get namespace: %w", err)
		}
		namespace = flag.Namespace(ns)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: load kubeclient configuration: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: load kubeclient configuration: %w", err)
	}

	dbInfo := &DBInfo{
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
		config:        kubeConfig,
		namespace:     namespace,
		appName:       appName,
	}

	isCloudSQL, err := IsCloudSQL(ctx, dbInfo)
	if err != nil {
		return nil, err
	}
	if isCloudSQL {
		return &CloudSQLDBInfo{
			DBInfo: dbInfo,
		}, nil
	} else {
		return NewPostgresDBInfo(ctx, dbInfo)
	}
}

func IsCloudSQL(ctx context.Context, i *DBInfo) (bool, error) {
	sqlInstances, err := i.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "sql.cnrm.cloud.google.com",
		Version:  "v1beta1",
		Resource: "sqlinstances",
	}).Namespace(string(i.namespace)).List(ctx, meta_v1.ListOptions{
		LabelSelector: "app=" + i.appName,
	})
	if err != nil {
		return false, fmt.Errorf("fetchDBInstance: error looking for sqlinstance %q in %q: %w", i.appName, i.namespace, err)
	}

	if len(sqlInstances.Items) == 1 {
		return true, nil
	} else if len(sqlInstances.Items) > 1 {
		return true, fmt.Errorf("fetchDBInstance: multiple sqlinstances found for app %q in %q", i.appName, i.namespace)
	}

	return false, nil
}

type ConnectionInfo struct {
	username string
	password string
	dbName   string
	instance string
	port     string
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
