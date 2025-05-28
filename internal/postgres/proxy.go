package postgres

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	"github.com/nais/cli/internal/output"
)

func RunProxy(ctx context.Context, appName, cluster, namespace, host string, port uint, verbose bool, out output.Output) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	projectID, err := dbInfo.ProjectID(ctx)
	if err != nil {
		return err
	}

	connectionName, err := dbInfo.ConnectionName(ctx)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%v:%v", host, port)

	out.Printf("Starting proxy on %v\n", address)
	out.Println("If you are using psql, you can connect to the database by running:")
	out.Printf("psql -h %v -p %v -U %v %v\n", host, port, email, connectionInfo.dbName)
	out.Println()
	out.Println("If you are using a JDBC client, you can connect to the database by using the following connection string:")
	out.Printf("Connection URL: jdbc:postgresql://%v/%v?user=%v\n", address, connectionInfo.dbName, email)
	out.Println()
	out.Println("If you get asked for a password, you can leave it blank. If that doesn't work, try running 'nais postgres grant", appName+"' again.")

	err = runProxy(ctx, projectID, connectionName, address, make(chan int, 1), verbose, out)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		fmt.Fprintln(os.Stderr, "\nERROR:", err)
	}

	return nil
}

func runProxy(ctx context.Context, projectID, connectionName, address string, port chan int, verbose bool, out output.Output) error {
	err := checkPostgresqlPassword(out)
	if err != nil {
		return err
	}

	if !verbose {
		logging.DisableLogging()
	}

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

	out.Println("Listening on", listener.Addr().String())

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

		out.Println("New connection", conn.RemoteAddr())
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
			out.Println("Connection complete", conn.RemoteAddr())
		}()
	}

	out.Println("Waiting for connections to close")
	wg.Wait()

	return nil
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}

func checkPostgresqlPassword(out output.Output) error {
	if _, ok := os.LookupEnv("PGPASSWORD"); ok {
		return fmt.Errorf("PGPASSWORD is set, please unset it before running this command")
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		out.Println("could not get home directory, can not check for .pgpass file")
		return nil
	}

	if s, err := os.Stat(filepath.Join(dirname, ".pgpass")); err == nil && !s.IsDir() {
		return fmt.Errorf("found .pgpass file in home directory, please remove it before running this command")
	}
	return nil
}
