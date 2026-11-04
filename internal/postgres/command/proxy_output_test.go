package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	for _, tt := range []struct {
		name          string
		env, password bool
	}{
		{name: "default human output"},
		{name: "env without password", env: true},
		{name: "env with password", env: true, password: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			previous := os.Stderr
			os.Stderr = stderr
			t.Cleanup(func() { os.Stderr = previous; _ = stderr.Close() })
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
					response = `{"data":{"team":{"environment":{"postgresAccess":{"connection":{"username":"user@example.com","password":"secret-password","caCertificate":"test-ca","serverName":"postgres.internal","relayEndpoint":"https://relay.example.com","relayAccess":"team/access","relayToken":"secret-relay-token"}}}}}}`
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
				}
				_, _ = fmt.Fprint(w, response)
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var output proxyOutputWriter
			var caPath, address string
			output.onWrite = func(text string) {
				if tt.env && strings.Contains(text, "PGCONNECT_TIMEOUT=") {
					var host, port string
					for line := range strings.SplitSeq(text, "\n") {
						key, value, _ := strings.Cut(line, "=")
						value = strings.Trim(value, "'")
						switch key {
						case "PGSSLROOTCERT":
							caPath = value
						case "PGHOSTADDR":
							host = value
						case "PGPORT":
							port = value
						}
					}
					address = net.JoinHostPort(host, port)
					content, err := os.ReadFile(caPath)
					if err != nil || string(content) != "test-ca" {
						t.Errorf("CA not ready at output time: %q, %v", content, err)
					}
					info, err := os.Stat(caPath)
					if err != nil || info.Mode().Perm() != 0o600 {
						t.Errorf("CA file must be owner-only: %v", err)
					}
					cancel()
				} else if !tt.env && strings.Contains(text, "Database password is only available") {
					cancel()
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
			args := []string{"--config", config, "postgres", "proxy", "orders", "-t", "my-team", "-e", "dev-gcp", "--branch", "main", "--reason", "Investigating issue"}
			if tt.env {
				args = append(args, "--output", "env")
			}
			if tt.password {
				args = append(args, "--print-password")
			}
			if err := app.Run(naistrix.RunWithArgs(args), naistrix.RunWithContext(ctx)); err != nil {
				t.Fatal(err)
			}
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatal("proxy did not print connection settings")
			}
			got := output.String()
			if tt.env {
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
			if tt.env && !strings.Contains(string(diagnostics), "Preparing personal Postgres access") {
				t.Errorf("missing progress on stderr: %s", diagnostics)
			}
			if tt.env && strings.Contains(string(diagnostics), "Warning: printing a database password") != tt.password {
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
