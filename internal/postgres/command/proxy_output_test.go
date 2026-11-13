package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type proxyOutputWriter struct {
	bytes.Buffer
	onWrite func(string)
}

func (w *proxyOutputWriter) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if w.onWrite != nil {
		w.onWrite(w.String())
	}
	return n, err
}

func TestProxyOutput(t *testing.T) {
	const password = "secret-password'\"$()\\= with spaces"
	for _, tt := range []struct {
		name, format        string
		password            bool
		exitOnEOF           bool
		closedStdin         bool
		cancelWithOpenStdin bool
	}{
		{name: "default human output"},
		{name: "env without password", format: "env"},
		{name: "env with password", format: "env", password: true},
		{name: "json without password", format: "json"},
		{name: "json with password", format: "json", password: true},
		{name: "env stops on stdin EOF", format: "env", exitOnEOF: true},
		{name: "json stops on stdin EOF", format: "json", exitOnEOF: true, password: true},
		{name: "human stops on stdin EOF", exitOnEOF: true},
		{name: "already closed stdin stops with flag", format: "json", exitOnEOF: true, closedStdin: true},
		{name: "closed stdin is ignored by default", format: "json", closedStdin: true},
		{name: "context cancellation still stops with open stdin", format: "json", exitOnEOF: true, cancelWithOpenStdin: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdout, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			previousStdout := os.Stdout
			os.Stdout = stdout
			t.Cleanup(func() { os.Stdout = previousStdout; _ = stdout.Close() })
			stdin, input, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			previousStdin := os.Stdin
			os.Stdin = stdin
			t.Cleanup(func() {
				os.Stdin = previousStdin
				_ = input.Close()
				_ = stdin.Close()
			})
			if tt.closedStdin {
				if err := input.Close(); err != nil {
					t.Fatal(err)
				}
			}
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			previous := os.Stderr
			os.Stderr = stderr
			t.Cleanup(func() { os.Stderr = previous; _ = stderr.Close() })
			config := setupProxyOutputAPI(t, password)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var stopTimer *time.Timer
			defer func() {
				if stopTimer != nil {
					stopTimer.Stop()
				}
			}()
			stopProxy := sync.OnceFunc(func() {
				if tt.exitOnEOF && !tt.cancelWithOpenStdin {
					if !tt.closedStdin {
						// Input data must be ignored, not mistaken for EOF.
						if _, err := input.WriteString("ignored input\n"); err != nil {
							t.Error(err)
						}
						if err := input.Close(); err != nil {
							t.Error(err)
						}
					}
				} else if tt.closedStdin {
					// With the flag disabled, EOF must not stop the proxy before
					// this explicit cancellation occurs.
					stopTimer = time.AfterFunc(100*time.Millisecond, cancel)
				} else {
					cancel()
				}
			})
			var output proxyOutputWriter
			var caPath, address string
			output.onWrite = func(text string) {
				if tt.format != "" {
					values := make(map[string]string)
					if tt.format == "json" {
						if err := json.Unmarshal([]byte(text), &values); err != nil {
							return // Wait until the complete object is written.
						}
					} else {
						if !strings.Contains(text, "PGCONNECT_TIMEOUT=") {
							return
						}
						for line := range strings.SplitSeq(text, "\n") {
							key, value, _ := strings.Cut(line, "=")
							values[key] = strings.Trim(value, "'")
						}
					}
					caPath = values["PGSSLROOTCERT"]
					address = net.JoinHostPort(values["PGHOSTADDR"], values["PGPORT"])
					content, err := os.ReadFile(caPath)
					if err != nil || string(content) != "test-ca" {
						t.Errorf("CA not ready at output time: %q, %v", content, err)
					}
					info, err := os.Stat(caPath)
					if err != nil || info.Mode().Perm() != 0o600 {
						t.Errorf("CA file must be owner-only: %v", err)
					}
					stopProxy()
				} else if strings.Contains(text, "Database password is only available") {
					stopProxy()
				}
			}
			app, global, err := naistrix.NewApplication("proxy-test", "Test proxy", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Postgres(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "postgres", "proxy", "orders", "-t", "my-team", "-e", "dev-gcp", "--reason", "Investigating issue"}
			if tt.format != "" {
				args = append(args, "--output", tt.format)
			}
			if tt.password {
				args = append(args, "--print-password")
			}
			if tt.exitOnEOF {
				args = append(args, "--exit-on-stdin-eof")
			}
			if err := app.Run(naistrix.RunWithArgs(args), naistrix.RunWithContext(ctx)); err != nil {
				t.Fatal(err)
			}
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatal("proxy did not print connection settings")
			}
			if tt.exitOnEOF && !tt.cancelWithOpenStdin && ctx.Err() != nil {
				t.Fatalf("expected stdin EOF to stop the proxy without canceling the parent context: %v", ctx.Err())
			}
			if !tt.exitOnEOF && tt.closedStdin && ctx.Err() == nil {
				t.Fatal("stdin EOF stopped the proxy without --exit-on-stdin-eof")
			}
			_, stdoutErr := stdout.Stat()
			if tt.format != "" && !errors.Is(stdoutErr, os.ErrClosed) {
				t.Errorf("machine-readable output must close stdout: %v", stdoutErr)
			} else if tt.format == "" && stdoutErr != nil {
				t.Errorf("human output must leave stdout open: %v", stdoutErr)
			}
			got := output.String()
			if tt.format != "" {
				if tt.format == "json" {
					var values map[string]string
					if err := json.Unmarshal([]byte(got), &values); err != nil {
						t.Fatalf("stdout must contain exactly one JSON object with string values: %v: %s", err, got)
					}
					want := map[string]string{
						"PGHOST": "postgres.internal", "PGHOSTADDR": "127.0.0.1", "PGUSER": "user@example.com",
						"PGSSLMODE": "verify-full", "PGDATABASE": "app", "PGCONNECT_TIMEOUT": "10", "PGSSLROOTCERT": caPath,
					}
					if tt.password {
						want["PGPASSWORD"] = password
					}
					for key, value := range want {
						if values[key] != value {
							t.Errorf("%s = %q, want %q", key, values[key], value)
						}
					}
					if values["PGPORT"] == "" || len(values) != len(want)+1 {
						t.Errorf("unexpected JSON keys: %v", values)
					}
				} else {
					for line := range strings.SplitSeq(strings.TrimSpace(got), "\n") {
						if !strings.HasPrefix(line, "PG") || !strings.Contains(line, "='") || !strings.HasSuffix(line, "'") {
							t.Errorf("non-assignment on stdout: %q", line)
						}
					}
					for _, want := range []string{"PGHOST='postgres.internal'", "PGHOSTADDR='127.0.0.1'", "PGUSER='user@example.com'", "PGSSLMODE='verify-full'", "PGDATABASE='app'"} {
						if !strings.Contains(got, want) {
							t.Errorf("stdout missing %q: %s", want, got)
						}
					}
				}
				if _, err := os.Stat(caPath); !os.IsNotExist(err) {
					t.Errorf("CA was not removed: %v", err)
				}
				listener, err := net.Listen("tcp", address)
				if err != nil {
					t.Errorf("listener was not closed: %v", err)
				} else {
					_ = listener.Close()
				}
			} else if !strings.Contains(got, "Postgres relay listening at 127.0.0.1:") {
				t.Errorf("missing human output: %s", got)
			}
			if strings.Contains(got, "secret-password") != tt.password {
				t.Errorf("password presence does not match --print-password: %s", got)
			}
			diagnostics, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if tt.format != "" && !strings.Contains(string(diagnostics), "Preparing personal Postgres access") {
				t.Errorf("missing progress on stderr: %s", diagnostics)
			}
			if tt.format != "" && strings.Contains(string(diagnostics), "Warning: printing a database password") != tt.password {
				t.Errorf("password warning does not match --print-password: %s", diagnostics)
			}
			for _, secret := range []string{"secret-relay-token", "https://relay.example.com", "team/access", "test-ca"} {
				if strings.Contains(got+string(diagnostics), secret) {
					t.Errorf("output contains relay connection material %q", secret)
				}
			}
		})
	}
}

func setupProxyOutputAPI(t *testing.T, password string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct{ OperationName string }
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		var response string
		switch request.OperationName {
		case "GetTeamPostgreses":
			response = `{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null}],"pageInfo":{"hasNextPage":false}}}}}`
		case "CreatePostgresAccess":
			response = `{"data":{"createPostgresAccess":{"name":"access"}}}`
		case "GetPostgresAccessStatus":
			response = `{"data":{"team":{"environment":{"postgresAccess":{"state":"READY"}}}}}`
		case "GetPostgresAccessConnection":
			encodedPassword, err := json.Marshal(password)
			if err != nil {
				t.Error(err)
			}
			response = fmt.Sprintf(`{"data":{"team":{"environment":{"postgresAccess":{"connection":{"username":"user@example.com","password":%s,"caCertificate":"test-ca","serverName":"postgres.internal","relayEndpoint":"https://relay.example.com","relayAccess":"team/access","relayToken":"secret-relay-token"}}}}}}`, encodedPassword)
		default:
			t.Errorf("unexpected operation: %s", request.OperationName)
		}
		_, _ = fmt.Fprint(w, response)
	}))
	t.Cleanup(server.Close)
	t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
	t.Setenv("HOME", t.TempDir())
	config := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return config
}

func TestProxyStdoutEOFBeforeShutdown(t *testing.T) {
	for _, format := range []string{"env", "json"} {
		t.Run(format, func(t *testing.T) {
			config := setupProxyOutputAPI(t, "secret-password")
			stdout, output, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = stdout.Close(); _ = output.Close() }()
			stdin, input, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = stdin.Close(); _ = input.Close() }()
			previousStdout, previousStdin := os.Stdout, os.Stdin
			os.Stdout, os.Stdin = output, stdin
			defer func() { os.Stdout, os.Stdin = previousStdout, previousStdin }()
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			previousStderr := os.Stderr
			os.Stderr = stderr
			defer func() { os.Stderr = previousStderr; _ = stderr.Close() }()

			app, global, err := naistrix.NewApplication("proxy-test", "Test proxy", "test", naistrix.ApplicationWithWriter(output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Postgres(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			args := []string{"--config", config, "postgres", "proxy", "orders", "-t", "my-team", "-e", "dev-gcp", "--reason", "Investigating issue", "--output", format, "--print-password", "--exit-on-stdin-eof"}
			var runErr error
			done := make(chan struct{})
			go func() {
				runErr = app.Run(naistrix.RunWithArgs(args), naistrix.RunWithContext(ctx))
				close(done)
			}()
			// Always join the command before restoring process-wide streams, including
			// when an assertion fails. Cancellation handles a missing stdout close.
			defer func() {
				cancel()
				_ = input.Close()
				<-done
				if runErr != nil {
					t.Errorf("proxy failed: %v", runErr)
				}
			}()
			if err := stdout.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(stdout)
			if err != nil {
				t.Fatalf("reading settings must reach EOF while proxy is running: %v", err)
			}
			values := make(map[string]string)
			if format == "json" {
				if err := json.Unmarshal(data, &values); err != nil {
					t.Fatalf("invalid JSON settings: %v: %s", err, data)
				}
			} else {
				for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
					key, value, ok := strings.Cut(line, "=")
					if !ok || !strings.HasPrefix(value, "'") || !strings.HasSuffix(value, "'") {
						t.Fatalf("invalid env assignment: %q", line)
					}
					values[key] = strings.Trim(value, "'")
				}
			}
			if values["PGDATABASE"] != "app" || values["PGPASSWORD"] != "secret-password" || values["PGSSLMODE"] != "verify-full" {
				t.Fatalf("incomplete settings at EOF: %v", values)
			}
			if ctx.Err() != nil {
				t.Fatalf("proxy did not remain running: %v", ctx.Err())
			}
			select {
			case <-done:
				t.Fatalf("proxy exited before stdin was closed: %v", runErr)
			default:
			}
			caPath := values["PGSSLROOTCERT"]
			if content, err := os.ReadFile(caPath); err != nil || string(content) != "test-ca" {
				t.Fatalf("CA must remain available after stdout EOF: %q, %v", content, err)
			}
			address := net.JoinHostPort(values["PGHOSTADDR"], values["PGPORT"])
			if listener, err := net.Listen("tcp", address); err == nil {
				_ = listener.Close()
				t.Fatal("proxy listener was closed along with stdout")
			}
			if err := input.Close(); err != nil {
				t.Fatal(err)
			}
			select {
			case <-done:
				if runErr != nil {
					t.Fatal(runErr)
				}
			case <-ctx.Done():
				t.Fatal("proxy did not stop after stdin EOF")
			}
			if _, err := os.Stat(caPath); !os.IsNotExist(err) {
				t.Errorf("CA was not removed on shutdown: %v", err)
			}
			listener, err := net.Listen("tcp", address)
			if err != nil {
				t.Errorf("listener was not closed on shutdown: %v", err)
			} else {
				_ = listener.Close()
			}
		})
	}
}
