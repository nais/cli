package postgres

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"strings"

	corev1 "k8s.io/api/core/v1"

	naisalpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

	multiDB      bool
	instanceName string
	databaseName string
	user         string
}

func NewDBInfo(appName, namespace, context, databaseName string) (*DBInfo, error) {
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
		databaseName:  databaseName,
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
	if err := i.fetchSQLDatabases(ctx); err != nil {
		return nil, err
	}

	if i.multiDB {
		i, err := i.dbConnectionMultiDB(ctx)
		if err != nil {
			return nil, err
		}

		return i, nil
	}

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

func (i *DBInfo) dbConnectionMultiDB(ctx context.Context) (*ConnectionInfo, error) {
	secrets, err := i.k8sClient.CoreV1().Secrets(i.namespace).List(ctx, v1.ListOptions{
		LabelSelector: "app=" + i.appName,
	})
	if err != nil {
		return nil, fmt.Errorf("unable list secrets for app %q in %q: %w", i.appName, i.namespace, err)
	}

	connectionName, err := i.ConnectionName(ctx)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.GetName(), "google-sql-"+i.appName+"-"+i.databaseName+"-"+i.user+"-") {
			return createConnectionInfo(secret, connectionName), nil
		}
	}

	return nil, fmt.Errorf("unable to find secret for app %q in %q with database %q and user %q", i.appName, i.namespace, i.databaseName, i.user)
}

func createConnectionInfo(secret corev1.Secret, instance string) *ConnectionInfo {
	return &ConnectionInfo{
		username: getSecretDataValue(secret, "_USERNAME"),
		password: getSecretDataValue(secret, "_PASSWORD"),
		dbName:   getSecretDataValue(secret, "_DATABASE"),
		port:     getSecretDataValue(secret, "_PORT"),
		host:     getSecretDataValue(secret, "_HOST"),
		instance: instance,
	}
}

func (i *DBInfo) fetchSQLDatabases(ctx context.Context) error {
	app := &naisalpha1.Application{}
	u, err := i.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "nais.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).Namespace(i.namespace).Get(ctx, i.appName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("fetchSQLDatabases: unable to get application %q in %q: %w", i.appName, i.namespace, err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, app); err != nil {
		return fmt.Errorf("fetchSQLDatabases: unable to convert unstructured to application: %w", err)
	}

	if app.Spec.GCP != nil && len(app.Spec.GCP.SqlInstances) != 1 {
		return fmt.Errorf("fetchSQLDatabases: expected exactly one sqlinstance, found %d", len(app.Spec.GCP.SqlInstances))
	}

	if len(app.Spec.GCP.SqlInstances[0].Databases) == 1 {
		return nil
	}

	if i.databaseName == "" {
		return fmt.Errorf("multiple databases found for app %q in %q, please specify one using the --database flag", i.appName, i.namespace)
	}

	for _, db := range app.Spec.GCP.SqlInstances[0].Databases {
		if db.Name == i.databaseName {
			i.multiDB = true
			i.instanceName = app.Spec.GCP.SqlInstances[0].Name
			i.user = db.Users[0].Name
			return nil
		}
	}

	return fmt.Errorf("database %q not found for app %q in %q", i.databaseName, i.appName, i.namespace)
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
	if !ok {
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
}

func (c *ConnectionInfo) ConnectionString() string {
	return fmt.Sprintf("host=%v user=%v dbname=%v password=%v sslmode=disable", c.instance, c.username, c.dbName, c.password)
}

func (c *ConnectionInfo) JDBCURL() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", c.username, c.password, c.host, c.port, c.dbName)
}

func (c *ConnectionInfo) SetPassword(password string) {
	c.password = password
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
			return fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth application-default login` first\n")
		}
	}

	return err
}
