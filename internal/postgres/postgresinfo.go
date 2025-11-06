package postgres

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/nais/naistrix"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type postgresDBInfo struct {
	*DBInfo
	clusterName string
}

func NewPostgresDBInfo(ctx context.Context, dbInfo *DBInfo) (DB, error) {
	p := &postgresDBInfo{
		DBInfo: dbInfo,
	}
	err := p.fetchClusterInfo(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *postgresDBInfo) DBConnection(ctx context.Context) (*ConnectionInfo, error) {
	email, err := currentEmail(ctx)
	if err != nil {
		return nil, err
	}

	queries := url.Values{}
	queries.Add("sslmode", "required")
	pgUrl := &url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(email, ""),
		Host:     "localhost",
		Path:     "app",
		RawQuery: queries.Encode(),
	}

	queries.Add("user", email)
	jdbcUrl := &url.URL{
		Scheme:   "jdbc:postgresql",
		Host:     "localhost:5432",
		Path:     "app",
		RawQuery: queries.Encode(),
	}

	return &ConnectionInfo{
		username: email,
		dbName:   "app",
		instance: "localhost",
		port:     "5432",
		url:      pgUrl,
		jdbcUrl:  jdbcUrl,
	}, nil
}

func (p *postgresDBInfo) RunProxy(ctx context.Context, host string, port *uint, portCh chan<- int, out *naistrix.OutputWriter, printInstructions bool) error {
	cfg, err := p.DBInfo.config.ClientConfig()
	if err != nil {
		return err
	}

	pods, err := p.DBInfo.k8sClient.CoreV1().Pods(fmt.Sprintf("pg-%s", p.namespace)).List(ctx, meta_v1.ListOptions{
		LabelSelector: fmt.Sprintf("spilo-role=master,application=spilo,cluster-name=%s", p.clusterName),
	})
	if err != nil {
		return err
	}
	if len(pods.Items) != 1 {
		return fmt.Errorf("found %d pods marked as master for cluster %s", len(pods.Items), p.clusterName)
	}

	masterPod := pods.Items[0]
	pfUrl := p.DBInfo.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(masterPod.GetNamespace()).
		Name(masterPod.GetName()).
		SubResource("portforward").
		URL()

	out.Verbosef("attempting port forward with URL: %s\n", pfUrl.String())
	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return errors.Wrap(err, "Could not create round tripper")
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", pfUrl)

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	ports := []string{":5432"}
	if port != nil {
		ports = []string{fmt.Sprintf("%d:5432", *port)}
	}

	out.Verbosef("Creating new portforward on %s for ports %v\n", host, ports)
	pf, err := portforward.NewOnAddresses(dialer, []string{host}, ports, stopChan, readyChan, NewNaisOut(out), NewNaisErr(out))
	if err != nil {
		return err
	}

	go func() {
		out.Verbosef("forwarding ports ...\n")
		errChan <- pf.ForwardPorts()
	}()

	out.Verbosef("Waiting for forwarding to be ready ...\n")
	select {
	case err = <-errChan:
		return errors.Wrap(err, "Could not create port forward")
	case <-readyChan:
	}

	if printInstructions {
		connectionInfo, err := p.DBConnection(ctx)
		if err != nil {
			return err
		}

		email, err := currentEmail(ctx)
		if err != nil {
			return err
		}

		out.Printf("Starting proxy on %s:%d\n", host, *port)
		out.Println()
		out.Println("Before you can connect, you need to request an access token:")
		out.Println("nais auth login --nais")
		out.Println("After logging in, you can get the current password using this command:")
		out.Println("nais auth print-access-token --nais")
		out.Println()
		out.Println("To connect to the database using psql, use the following command:")
		out.Printf("PGPASSWORD=$(nais auth print-access-token --nais) psql -h %v -p %v -U %v %v\n", host, *port, email, connectionInfo.dbName)
		out.Println()
		out.Println("If you are using a JDBC client, you can connect to the database by using the following connection string:")
		out.Printf("Connection URL: %s\n", connectionInfo.jdbcUrl)
	}

	forwardedPorts, err := pf.GetPorts()
	if err != nil {
		return err
	}
	for _, forwardedPort := range forwardedPorts {
		out.Infof("Listening on %s:%d\n", host, forwardedPort.Local)
		portCh <- int(forwardedPort.Local)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errChan:
		return errors.Wrap(err, "Could not create port forward")
	}
}

func (p *postgresDBInfo) ToCloudSQLDBInfo() (*CloudSQLDBInfo, error) {
	return nil, fmt.Errorf("not a CloudSQL instance")
}

func (p *postgresDBInfo) fetchClusterInfo(ctx context.Context) error {
	unstructuredApp, err := p.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "nais.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).Namespace(string(p.namespace)).Get(ctx, p.appName, meta_v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("fetchClusterInfo: error looking for Application %q in %q: %w", p.appName, p.namespace, err)
	}

	app := &nais_io_v1alpha1.Application{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredApp.Object, app)
	if err != nil {
		return fmt.Errorf("fetchClusterInfo: error converting to Application %q in %q: %w", p.appName, p.namespace, err)
	}

	if app.Spec.Postgres == nil {
		return fmt.Errorf("fetchClusterInfo: application %q in %q does not have a Postgres cluster", p.appName, p.namespace)
	}

	p.clusterName = app.Spec.Postgres.ClusterName

	return nil
}

func NewNaisOut(out *naistrix.OutputWriter) io.Writer {
	return &NaisWriter{
		writeFunc: out.Infof,
	}
}

func NewNaisErr(out *naistrix.OutputWriter) io.Writer {
	return &NaisWriter{
		writeFunc: out.Errorf,
	}
}

type NaisWriter struct {
	writeFunc func(string, ...any)
}

func (o *NaisWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	o.writeFunc(msg)
	return len(msg), nil
}
