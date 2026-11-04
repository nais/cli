package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/relay"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

func psqlCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.Access{
		Postgres:    parent,
		AccessLevel: "read",
		TTL:         30 * time.Minute,
		Database:    "app",
	}
	return &naistrix.Command{
		Name:              "psql",
		Title:             "Connect to Nais Postgres via psql.",
		Description:       "Request personal access, open a local relay tunnel and start psql with end-to-end TLS verification. If --reason is omitted, prompt for an audit reason interactively.",
		Args:              []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance to connect to", ChoicesFunc: postgresNameChoices(parent)}},
		Flags:             f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			report := func(message string) { out.Println(message) }
			request, err := prepareAccessRequest(args.Get("postgres"), f, func(message string) { out.Warnln(message) })
			if err != nil {
				return err
			}
			// SIGTERM must run the deferred cleanup (CA file, relay). Ctrl-C aborts the setup, but once
			// psql runs it is left to psql (query cancellation) and ignored here until cleanup is done.
			ctx, stopTerm := signal.NotifyContext(ctx, syscall.SIGTERM)
			defer stopTerm()
			setupCtx, stopSetup := signal.NotifyContext(ctx, os.Interrupt)
			connection, err := requestAccess(setupCtx, request, report)
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
			cmd.Env = append(withoutPostgresEnv(os.Environ()), postgresConnectionEnv(connection, listener.Addr().(*net.TCPAddr), ca.Name(), true)...)
			cmd.Env = append(cmd.Env, "PGDATABASE="+f.Database)
			out.Println("Connecting with verified PostgreSQL TLS through the local relay...")
			return cmd.Run()
		},
	}
}

func proxyCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.Proxy{
		Postgres:    parent,
		AccessLevel: "read",
		TTL:         30 * time.Minute, Host: "127.0.0.1",
		Output: "human",
	}
	return &naistrix.Command{
		Name: "proxy", Title: "Expose a Nais Postgres relay tunnel locally.",
		Description: heredoc.Doc(`
			Request personal access and listen on loopback. If <info>--reason</info> is omitted, prompt for an audit reason interactively.

			PostgreSQL clients must verify the server certificate. Use <info>nais postgres psql</info> for automatic TLS setup.

			Use <info>--output env</info> for POSIX shell-quoted PostgreSQL environment variable assignments on stdout.
			Use <info>--output json</info> for a JSON object with the same variable names as keys and string values, without shell quoting. Both formats send progress and warnings to stderr.
			<info>PGHOST</info> is the TLS server name; <info>PGHOSTADDR</info> and <info>PGPORT</info> point to the local relay.
			<info>PGDATABASE</info> is always <info>app</info>, the application database. <info>PGPASSWORD</info> is omitted unless <info>--print-password</info> is set.

			Keep the proxy running while using these settings. The temporary <info>PGSSLROOTCERT</info> file is removed when it stops.

			Source the <info>env</info> output with auto-export enabled (<info>set -a</info>) to pass the variables to child processes.
			Programs consuming <info>json</info> output should decode the object as soon as it is available, rather than waiting for EOF; the proxy keeps running until stopped.
			Inherited PostgreSQL settings such as <info>PGSERVICE</info> may need to be unset.
		`),
		Examples: []naistrix.CommandExample{
			{Description: "Write connection settings to a file while the proxy runs. In another terminal, use 'set -a; . ./postgres.env; set +a' to export them. Use a private file (umask 077) when adding --print-password.", Command: "my-postgres -t my-team --reason 'Investigating issue' --output env > postgres.env"},
			{Description: "Emit connection settings as JSON for a program to consume. Add --print-password only when the program needs the password.", Command: "my-postgres -t my-team --reason 'Investigating issue' --output json"},
		},
		Args:              []naistrix.Argument{{Name: "postgres", Prompt: "Name of the Postgres instance to proxy", ChoicesFunc: postgresNameChoices(parent)}},
		Flags:             f,
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if f.Output != "human" && f.Output != "env" && f.Output != "json" {
				return fmt.Errorf("unsupported output format %q: expected human, env, or json", f.Output)
			}
			report := func(message string) { out.Println(message) }
			warn := func(message string) { out.Warnln(message) }
			if f.Output != "human" {
				report = func(message string) { _, _ = fmt.Fprintln(os.Stderr, message) }
				warn = report
			}
			if net.ParseIP(f.Host) == nil || !net.ParseIP(f.Host).IsLoopback() {
				return fmt.Errorf("--host must be a loopback IP address")
			}
			if f.Port < 0 || f.Port > 65535 {
				return fmt.Errorf("--port must be between 0 and 65535")
			}
			request, err := prepareAccessRequest(args.Get("postgres"), &flag.Access{
				Postgres: f.Postgres, Branch: f.Branch, AccessLevel: f.AccessLevel, Reason: f.Reason, TTL: f.TTL,
			}, warn)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
			connection, err := requestAccess(ctx, request, report)
			if err != nil {
				return err
			}
			listener, err := net.Listen("tcp", net.JoinHostPort(f.Host, fmt.Sprint(f.Port)))
			if err != nil {
				return err
			}
			defer func() { _ = listener.Close() }()
			ca, err := os.CreateTemp("", "nais-postgres-ca-*.crt")
			if err != nil {
				return err
			}
			defer func() { _ = os.Remove(ca.Name()) }()
			if _, err := ca.WriteString(connection.CACertificate); err != nil {
				_ = ca.Close()
				return err
			}
			if err := ca.Close(); err != nil {
				return err
			}
			if f.Output != "human" {
				if f.PrintPassword {
					report("Warning: printing a database password; protect the output and avoid terminal capture and shell history.")
				}
				env := postgresConnectionEnv(connection, listener.Addr().(*net.TCPAddr), ca.Name(), f.PrintPassword)
				env = append(env, "PGDATABASE=app")
				if f.Output == "json" {
					values := make(map[string]string, len(env))
					for _, entry := range env {
						key, value, _ := strings.Cut(entry, "=")
						values[key] = value
					}
					if err := out.JSON().Render(values); err != nil {
						return err
					}
				} else {
					out.Printf("%s", postgresEnvOutput(env))
				}
			} else {
				out.Printf("Postgres relay listening at %s; user: %s; TLS server name: %s; CA: %s\n", listener.Addr(), connection.Username, connection.ServerName, ca.Name())
				out.Println("Use host=<server name> hostaddr=<loopback IP> sslmode=verify-full sslrootcert=<CA path>. For automatic setup use 'nais postgres psql'.")
				out.Println("Database password is only available here with --print-password (sensitive output).")
				if f.PrintPassword {
					out.Errorf("Warning: printing a database password; avoid terminal capture and shell history.\n")
					out.Printf("Database password (sensitive): %s\n", connection.Password)
				}
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

func resolveAccessReason(provided string, report func(string)) (string, error) {
	reason := strings.TrimSpace(provided)
	if reason == "" {
		report("Requesting personal Postgres access is logged for auditing purposes.")
		entered, err := input.Input("Reason for personal Postgres access (min 10 chars)")
		if errors.Is(err, input.ErrNotInteractive) {
			return "", fmt.Errorf("missing required reason; specify --reason with at least 10 characters when running noninteractively")
		}
		if err != nil {
			return "", fmt.Errorf("prompting for access reason: %w", err)
		}
		reason = strings.TrimSpace(entered)
	}
	if len(reason) < 10 {
		return "", fmt.Errorf("--reason must contain at least 10 characters")
	}
	return reason, nil
}

// Prepare the audit reason before installing signal handlers: input.Input cannot observe context cancellation.
func prepareAccessRequest(name string, f *flag.Access, report func(string)) (gql.CreatePostgresAccessInput, error) {
	if f.TTL < time.Second || f.TTL > time.Hour {
		return gql.CreatePostgresAccessInput{}, fmt.Errorf("--ttl must be between 1s and 1h")
	}
	levels := map[flag.AccessLevel]gql.PostgresAccessLevel{
		"read":     gql.PostgresAccessLevelRead,
		"write":    gql.PostgresAccessLevelReadwrite,
		"admin":    gql.PostgresAccessLevelAdmin,
		"maintain": gql.PostgresAccessLevelMaintain,
	}
	level, ok := levels[f.AccessLevel]
	if !ok {
		return gql.CreatePostgresAccessInput{}, fmt.Errorf("--access-level must be read, write, admin, or maintain")
	}
	reason, err := resolveAccessReason(f.Reason, report)
	if err != nil {
		return gql.CreatePostgresAccessInput{}, err
	}
	ttl := f.TTL.String()
	return gql.CreatePostgresAccessInput{
		Postgres: name, Branch: f.Branch, TeamSlug: f.Team, EnvironmentName: string(f.Environment),
		AccessLevel: level, Reason: reason, Ttl: &ttl,
	}, nil
}

func requestAccess(ctx context.Context, request gql.CreatePostgresAccessInput, report func(string)) (postgres.Connection, error) {
	environment, err := resolvePostgresEnvironment(ctx, request.TeamSlug, request.Postgres, request.EnvironmentName)
	if err != nil {
		return postgres.Connection{}, err
	}
	var spinner *accessSpinner
	if term.IsTerminal(int(os.Stderr.Fd())) {
		spinner = newAccessSpinner(os.Stderr)
		defer func() {
			if spinner != nil {
				spinner.Stop()
			}
		}()
	} else {
		report("Preparing personal Postgres access...")
	}
	api, err := postgres.NewAPI(ctx)
	if err != nil {
		return postgres.Connection{}, err
	}
	request.EnvironmentName = environment
	if request.Branch == "" {
		lookupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		request.Branch, err = api.ActiveBranch(lookupCtx, request.TeamSlug, environment, request.Postgres)
		if err != nil {
			return postgres.Connection{}, err
		}
	}
	connection, err := postgres.CreateAndWait(ctx, api, request)
	if err == nil && spinner != nil {
		spinner.Stop()
		spinner = nil
		report("Postgres access ready.")
	}
	return connection, err
}
