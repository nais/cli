package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
	"golang.org/x/oauth2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DB interface {
	DBConnection(ctx context.Context) (*ConnectionInfo, error)
	RunProxy(ctx context.Context, host string, port *uint, portCh chan<- int, verbose bool, out *naistrix.OutputWriter) error

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

func NewDBInfo(appName string, namespace flag.Namespace, context flag.Context) (DB, error) {
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

	return &CloudSQLDBInfo{
		DBInfo: DBInfo{
			k8sClient:     k8sClient,
			dynamicClient: dynamicClient,
			config:        kubeConfig,
			namespace:     namespace,
			appName:       appName,
		},
	}, nil
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
