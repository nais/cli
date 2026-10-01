package command

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/alpha/postgres/relay"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func accessFlags(parent *flag.Postgres) *flag.Access {
	return &flag.Access{Postgres: parent, AccessLevel: "read", TTL: 30 * time.Minute, Database: "app"}
}

func psqlCommand(parent *flag.Postgres) *naistrix.Command {
	f := accessFlags(parent)
	return &naistrix.Command{
		Name: "psql", Title: "Connect to Nais Postgres via psql (experimental).",
		Description: "Request personal access, open a local relay tunnel and start psql with end-to-end TLS verification.",
		Args:        []naistrix.Argument{{Name: "postgres"}}, Flags: f,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			// SIGTERM must run the deferred cleanup (CA file, relay). Ctrl-C aborts the setup, but once
			// psql runs it is left to psql (query cancellation) and ignored here until cleanup is done.
			ctx, stopTerm := signal.NotifyContext(ctx, syscall.SIGTERM)
			defer stopTerm()
			setupCtx, stopSetup := signal.NotifyContext(ctx, os.Interrupt)
			connection, err := requestAccess(setupCtx, args.Get("postgres"), f)
			// Register the ignore-channel before releasing the setup handler so there is no unhandled window.
			interrupts := make(chan os.Signal, 1)
			signal.Notify(interrupts, os.Interrupt)
			defer signal.Stop(interrupts)
			stopSetup()
			if err != nil {
				return err
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return err
			}
			tunnelCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- relay.Serve(tunnelCtx, listener, relay.Tunnel{Endpoint: connection.RelayEndpoint, Access: connection.RelayAccess, Token: connection.RelayToken})
			}()
			defer func() { cancel(); <-done }()
			ca, err := os.CreateTemp("", "nais-postgres-ca-*.crt")
			if err != nil {
				return err
			}
			defer func() { _ = os.Remove(ca.Name()) }()
			defer func() { _ = ca.Close() }()
			if err := ca.Chmod(0o600); err != nil {
				return err
			}
			if _, err := ca.WriteString(connection.CACertificate); err != nil {
				return err
			}
			if err := ca.Close(); err != nil {
				return err
			}
			path, err := exec.LookPath("psql")
			if err != nil {
				return fmt.Errorf("psql not found: %w", err)
			}
			cmd := exec.CommandContext(ctx, path, "-X", "-w")
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			cmd.Env = append(withoutPostgresEnv(os.Environ()),
				"PGHOST="+connection.ServerName, "PGHOSTADDR=127.0.0.1", fmt.Sprintf("PGPORT=%d", listener.Addr().(*net.TCPAddr).Port),
				"PGUSER="+connection.Username, "PGPASSWORD="+connection.Password, "PGDATABASE="+f.Database,
				"PGSSLMODE=verify-full", "PGSSLROOTCERT="+ca.Name(), "PGCONNECT_TIMEOUT=10")
			out.Println("Connecting with verified PostgreSQL TLS through the local relay...")
			return cmd.Run()
		},
	}
}

func proxyCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.Proxy{Postgres: parent, AccessLevel: "read", TTL: 30 * time.Minute, Host: "127.0.0.1"}
	return &naistrix.Command{
		Name: "proxy", Title: "Expose a Nais Postgres relay tunnel locally (experimental).",
		Description: "Request personal access and listen on loopback. PostgreSQL clients must verify the server certificate; use psql for automatic TLS setup. Credentials are not printed unless --print-password is set.",
		Args:        []naistrix.Argument{{Name: "postgres"}}, Flags: f,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			if net.ParseIP(f.Host) == nil || !net.ParseIP(f.Host).IsLoopback() {
				return fmt.Errorf("--host must be a loopback IP address")
			}
			if f.Port < 0 || f.Port > 65535 {
				return fmt.Errorf("--port must be between 0 and 65535")
			}
			connection, err := requestAccess(ctx, args.Get("postgres"), &flag.Access{
				Postgres: f.Postgres, Branch: f.Branch, AccessLevel: f.AccessLevel, Reason: f.Reason, TTL: f.TTL,
			})
			if err != nil {
				return err
			}
			listener, err := net.Listen("tcp", net.JoinHostPort(f.Host, fmt.Sprint(f.Port)))
			if err != nil {
				return err
			}
			ca, err := os.CreateTemp("", "nais-postgres-ca-*.crt")
			if err != nil {
				_ = listener.Close()
				return err
			}
			defer func() { _ = os.Remove(ca.Name()) }()
			if _, err := ca.WriteString(connection.CACertificate); err != nil {
				_ = ca.Close()
				_ = listener.Close()
				return err
			}
			if err := ca.Close(); err != nil {
				_ = listener.Close()
				return err
			}
			out.Printf("Postgres relay listening at %s; user: %s; TLS server name: %s; CA: %s\n", listener.Addr(), connection.Username, connection.ServerName, ca.Name())
			out.Println("Use host=<server name> hostaddr=<loopback IP> sslmode=verify-full sslrootcert=<CA path>. For automatic setup use 'nais alpha postgres psql'.")
			out.Println("Database password is only available here with --print-password (sensitive output).")
			if f.PrintPassword {
				out.Errorf("Warning: printing a database password; avoid terminal capture and shell history.\n")
				out.Printf("Database password (sensitive): %s\n", connection.Password)
			}
			return relay.Serve(ctx, listener, relay.Tunnel{Endpoint: connection.RelayEndpoint, Access: connection.RelayAccess, Token: connection.RelayToken})
		},
	}
}

func withoutPostgresEnv(env []string) []string {
	ret := make([]string, 0, len(env))
	for _, item := range env {
		if !strings.HasPrefix(item, "PG") {
			ret = append(ret, item)
		}
	}
	return ret
}

func requestAccess(ctx context.Context, name string, f *flag.Access) (postgres.Connection, error) {
	if f.Team == "" || f.Environment == "" {
		return postgres.Connection{}, fmt.Errorf("--team and --environment are required")
	}
	if len(strings.TrimSpace(f.Reason)) < 10 {
		return postgres.Connection{}, fmt.Errorf("--reason must contain at least 10 characters")
	}
	if f.TTL < time.Second || f.TTL > time.Hour {
		return postgres.Connection{}, fmt.Errorf("--ttl must be between 1s and 1h")
	}
	levels := map[string]gql.PostgresAccessLevel{"read": gql.PostgresAccessLevelRead, "write": gql.PostgresAccessLevelReadwrite, "admin": gql.PostgresAccessLevelReadwritecreate}
	level, ok := levels[f.AccessLevel]
	if !ok {
		return postgres.Connection{}, fmt.Errorf("--access-level must be read, write, or admin")
	}
	api, err := postgres.NewAPI(ctx)
	if err != nil {
		return postgres.Connection{}, err
	}
	branch := f.Branch
	if branch == "" {
		lookupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		branch, err = api.ActiveBranch(lookupCtx, f.Team, string(f.Environment), name)
		if err != nil {
			return postgres.Connection{}, err
		}
	}
	ttl := f.TTL.String()
	return postgres.CreateAndWait(ctx, api, gql.CreatePostgresAccessInput{
		Postgres: name, Branch: branch, TeamSlug: f.Team, EnvironmentName: string(f.Environment),
		AccessLevel: level, Reason: f.Reason, Ttl: &ttl,
	})
}
