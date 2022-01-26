package postgres

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/certs"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"github.com/nais/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goauth "golang.org/x/oauth2/google"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy [app-name] [flags]",
	Short: "Create a proxy to a Postgres database",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()
		appName := args[0]
		namespace := viper.GetString(cmd.NamespaceFlag)
		context := viper.GetString(cmd.ContextFlag)
		port := viper.GetString(cmd.PortFlag)

		dbInfo, err := NewDBInfo(appName, namespace, context)
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

		return runProxy(ctx, projectID, connectionName, "localhost:"+port, make(chan int, 1))
	},
}

func runProxy(ctx context.Context, projectID, connectionName, address string, port chan int) error {
	if err := grantUserAccess(ctx, projectID, 1*time.Hour); err != nil {
		return err
	}

	fmt.Println("Proxy to instance", connectionName)
	if err := proxy.InitDefault(ctx); err != nil {
		return err
	}

	oauthClient, err := goauth.DefaultClient(ctx, "https://www.googleapis.com/auth/sqlservice.admin")
	if err != nil {
		return err
	}
	proxyClient := &proxy.Client{
		Port:  proxy.DefaultPort,
		Certs: certs.NewCertSource("", oauthClient, true),
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	connSrc := make(chan proxy.Conn, 2)
	// TODO(thokra): Make port configurable
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	port <- listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			c, err := listener.AcceptTCP()
			if err != nil {
				log.Println("Accept TCP", err)
				return
			}
			connSrc <- proxy.Conn{
				Instance: connectionName,
				Conn:     c,
			}
		}
	}()

	termTimeout := time.Second * 5
	go func() {
		defer func() { cancel() }()
		<-ctx.Done()

		log.Printf("Received TERM signal. Waiting up to %s before terminating.", termTimeout)
		if err := listener.Close(); err != nil {
			log.Println(err)
		}

		err := proxyClient.Shutdown(termTimeout)
		if err != nil {
			log.Printf("Error during SIGTERM shutdown: %v", err)
			return
		}
	}()

	proxyClient.RunContext(ctx, connSrc)

	return nil
}

func grantUserAccess(ctx context.Context, projectID string, duration time.Duration) error {
	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	args := []string{
		"projects",
		"add-iam-policy-binding",
		projectID,
		"--member", "user:" + email,
		"--role", "roles/cloudsql.instanceUser",
	}

	if duration > 0 {
		timestamp := time.Now().Add(duration).UTC().Format(time.RFC3339)
		args = append(args,
			"--condition",
			"expression=request.time < timestamp('"+timestamp+"'),title=temp_access",
		)
	}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
