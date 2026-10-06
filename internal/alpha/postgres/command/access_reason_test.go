package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestPersonalAccessReason(t *testing.T) {
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() {
		os.Stdin = previous
		if err := stdin.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, command := range []string{"psql", "proxy"} {
		t.Run(command, func(t *testing.T) {
			for _, tt := range []struct {
				name, wantError string
				args            []string
				wantAudit       bool
				wantMutation    bool
				proxyOnly       bool
				omitTeam        bool
			}{
				{name: "missing reason requires flag noninteractively", wantError: "specify --reason", wantAudit: true},
				{name: "whitespace reason prompts rather than being accepted", args: []string{"--reason", " \t "}, wantError: "specify --reason", wantAudit: true},
				{name: "short explicit reason is not reprompted", args: []string{"--reason", "short"}, wantError: "--reason must contain at least 10 characters"},
				{name: "trimmed explicit reason must be long enough", args: []string{"--reason", "     short     "}, wantError: "--reason must contain at least 10 characters"},
				{name: "team is required before prompting", omitTeam: true, wantError: "missing required team"},
				{name: "invalid proxy host fails before prompting", proxyOnly: true, args: []string{"--host", "0.0.0.0"}, wantError: "--host must be a loopback IP address"},
				{name: "invalid proxy port fails before prompting", proxyOnly: true, args: []string{"--port", "65536"}, wantError: "--port must be between 0 and 65535"},
				{name: "invalid lifetime fails before prompting", args: []string{"--ttl", "2h"}, wantError: "--ttl must be between 1s and 1h"},
				{name: "invalid access level fails before prompting", args: []string{"--access-level", "invalid"}, wantError: "--access-level must be read, write, or admin"},
				{name: "explicit reason bypasses prompt and is sent trimmed", args: []string{"--reason", "  Investigating issue  "}, wantError: "test stops after audited request", wantMutation: true},
			} {
				if tt.proxyOnly && command != "proxy" {
					continue
				}
				t.Run(tt.name, func(t *testing.T) {
					var operations []string
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var request struct {
							OperationName string `json:"operationName"`
							Variables     struct {
								Input gql.CreatePostgresAccessInput `json:"input"`
							} `json:"variables"`
						}
						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
							t.Errorf("decoding request: %v", err)
						}
						operations = append(operations, request.OperationName)
						if request.OperationName != "CreatePostgresAccess" || !tt.wantMutation {
							t.Errorf("unexpected API operation: %s", request.OperationName)
						}
						input := request.Variables.Input
						if input.Reason != "Investigating issue" || input.Postgres != "orders" || input.Branch != "main" || input.TeamSlug != "my-team" || input.EnvironmentName != "dev-gcp" || input.AccessLevel != gql.PostgresAccessLevelRead || input.Ttl == nil || *input.Ttl != "30m0s" {
							t.Errorf("incorrect audited access request: %+v", input)
						}
						w.Header().Set("Content-Type", "application/json")
						if _, err := w.Write([]byte(`{"errors":[{"message":"test stops after audited request"}]}`)); err != nil {
							t.Errorf("writing response: %v", err)
						}
					}))
					defer server.Close()
					t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
					t.Setenv("HOME", t.TempDir())
					config := filepath.Join(t.TempDir(), "config.yaml")
					if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					var output bytes.Buffer
					app, global, err := naistrix.NewApplication("postgres-access-test", "Test personal access", "test", naistrix.ApplicationWithWriter(&output))
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
					args := []string{"--config", config, "postgres", command, "orders", "-e", "dev-gcp", "--branch", "main"}
					if !tt.omitTeam {
						args = append(args, "-t", "my-team")
					}
					args = append(args, tt.args...)
					err = app.Run(naistrix.RunWithArgs(args))
					if err == nil || !strings.Contains(err.Error(), tt.wantError) {
						t.Errorf("error = %v; want %q", err, tt.wantError)
					}
					if got := strings.Contains(output.String(), "logged for auditing purposes"); got != tt.wantAudit {
						t.Errorf("audit prompt warning = %v, want %v: %s", got, tt.wantAudit, output.String())
					}
					if tt.wantMutation {
						if len(operations) != 1 || operations[0] != "CreatePostgresAccess" {
							t.Errorf("API operations = %v; want only the access request", operations)
						}
					} else if len(operations) != 0 {
						t.Errorf("API calls occurred before reason/flag validation: %v", operations)
					}
				})
			}
		})
	}
}
