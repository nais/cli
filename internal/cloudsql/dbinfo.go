package cloudsql

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/oauth2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DBInfo struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        clientcmd.ClientConfig
	namespace     string
	appName       string
}

func (d *DBInfo) AppName() string { return d.appName }

func NewDBInfo(_ context.Context, appName, team, environment string) (*CloudSQLDBInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{CurrentContext: environment})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("NewDBInfo: unable to get kubeconfig: %w", err)
	}
	if team == "" {
		team, _, err = kubeConfig.Namespace()
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
	return &CloudSQLDBInfo{DBInfo: &DBInfo{k8sClient: k8sClient, dynamicClient: dynamicClient, config: kubeConfig, namespace: team, appName: appName}}, nil
}

type ConnectionInfo struct {
	username string
	email    string
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
		c.jdbcUrl = &url.URL{Scheme: "jdbc:postgresql", Host: c.url.Host, Path: c.dbName, RawQuery: queries.Encode()}
	}
}

func formatInvalidGrantError(err error) error {
	var retrieve *oauth2.RetrieveError
	if errors.As(err, &retrieve) && retrieve.ErrorCode == "invalid_grant" {
		return fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth login --update-adc` first")
	}
	return err
}
