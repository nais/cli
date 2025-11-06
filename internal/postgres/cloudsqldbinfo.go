package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	"github.com/nais/naistrix"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

type CloudSQLDBInfo struct {
	*DBInfo
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
	secret, err := i.k8sClient.CoreV1().Secrets(string(i.namespace)).Get(ctx, "google-sql-"+i.appName, meta_v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get database password from %q in %q: %w", "google-sql-"+i.appName, i.namespace, err)
	}

	connectionName, err := i.ConnectionName(ctx)
	if err != nil {
		return nil, err
	}

	return createConnectionInfo(*secret, connectionName), nil
}

func createConnectionInfo(secret core_v1.Secret, instance string) *ConnectionInfo {
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

func getSecretDataValue(secret core_v1.Secret, suffix string) string {
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
	}).Namespace(string(i.namespace)).List(ctx, meta_v1.ListOptions{
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

func (d *CloudSQLDBInfo) RunProxy(ctx context.Context, host string, port *uint, portCh chan<- int, out *naistrix.OutputWriter, printInstructions bool) error {
	projectID, err := d.ProjectID(ctx)
	if err != nil {
		return err
	}

	connectionName, err := d.ConnectionName(ctx)
	if err != nil {
		return err
	}

	if port == nil {
		port = ptr.To(uint(0))
	}
	address := fmt.Sprintf("%s:%d", host, *port)

	if printInstructions {
		connectionInfo, err := d.DBConnection(ctx)
		if err != nil {
			return err
		}

		email, err := currentEmail(ctx)
		if err != nil {
			return err
		}

		out.Printf("Starting proxy on %v\n", address)
		out.Println("If you are using psql, you can connect to the database by running:")
		out.Printf("psql -h %v -p %d -U %v %v\n", host, *port, email, connectionInfo.dbName)
		out.Println()
		out.Println("If you are using a JDBC client, you can connect to the database by using the following connection string:")
		out.Printf("Connection URL: %v\n", connectionInfo.jdbcUrl)
		out.Println()
		out.Println("If you get asked for a password, you can leave it blank. If that doesn't work, try running 'nais postgres grant", d.AppName()+"' again.")
	}

	err = runProxy(ctx, projectID, connectionName, address, portCh, out)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		fmt.Fprintln(os.Stderr, "\nERROR:", err)
	}

	return nil
}

func runProxy(ctx context.Context, projectID, connectionName, address string, port chan<- int, out *naistrix.OutputWriter) error {
	err := checkPostgresqlPassword(out)
	if err != nil {
		return err
	}

	logging.Verbosef = out.Verbosef
	logging.Infof = out.Infof
	logging.Errorf = out.Errorf

	if err := grantUserAccess(ctx, projectID, "roles/cloudsql.instanceUser", 1*time.Hour, out); err != nil {
		return err
	}

	opts := []cloudsqlconn.Option{
		cloudsqlconn.WithIAMAuthN(),
	}
	d, err := cloudsqlconn.NewDialer(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create dialer: %w", err)
	}

	if err := d.Warmup(ctx, connectionName); err != nil {
		return fmt.Errorf("failed to warmup connection: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP address: %w", err)
	}

	out.Infof("Listening on %s\n", listener.Addr().String())

	port <- listener.Addr().(*net.TCPAddr).Port

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			out.Println("error closing listener", err)
		}
	}()

	wg := sync.WaitGroup{}
	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			out.Println("error accepting connection", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		out.Infof("New connection %s\n", conn.RemoteAddr())
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				<-ctx.Done()
				if err := conn.Close(); err != nil {
					out.Println("error closing connection", err)
				}
			}()

			conn2, err := d.Dial(ctx, connectionName)
			if err != nil {
				out.Println("error dialing connection", err)
				return
			}
			defer conn2.Close()

			closer := make(chan struct{}, 2)
			go copy(closer, conn2, conn)
			go copy(closer, conn, conn2)
			<-closer
			out.Infof("Connection complete %s", conn.RemoteAddr())
		}()
	}

	out.Infof("Waiting for connections to close\n")
	wg.Wait()

	return nil
}
