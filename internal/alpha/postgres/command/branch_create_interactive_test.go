package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestBranchCreateNameAndProgress(t *testing.T) {
	for _, tt := range []struct {
		name, config, returnedName, apiError, wantError string
		positionals, extraFlags                         []string
		omitTeam, omitTime, resolveEnvironment          bool
		wantOperations                                  []string
	}{
		{
			name: "explicit branch bypasses prompt", positionals: []string{"orders", "5m-restore"}, returnedName: "5m-returned",
			wantOperations: []string{"CreatePostgresBranch"},
		},
		{
			name: "default environment", positionals: []string{"orders", "5m-restore"}, returnedName: "5m-returned", config: "environment: dev-gcp\n", resolveEnvironment: true,
			wantOperations: []string{"CreatePostgresBranch"},
		},
		{
			name: "explicit environment overrides default", positionals: []string{"orders", "5m-restore"}, returnedName: "5m-returned", config: "environment: prod-gcp\n",
			wantOperations: []string{"CreatePostgresBranch"},
		},
		{
			name: "resolved environment", positionals: []string{"orders", "5m-restore"}, returnedName: "5m-returned", resolveEnvironment: true,
			wantOperations: []string{"GetTeamPostgreses", "CreatePostgresBranch"},
		},
		{
			name: "missing branch requires positional argument noninteractively", positionals: []string{"orders"}, resolveEnvironment: true,
			wantError: "specify the new branch name as the second positional argument, e.g. nais alpha postgres branch create mypg 5m-restore --ago 5m",
		},
		{name: "no Postgres", wantError: "Expected at least 1 argument"},
		{name: "empty Postgres", positionals: []string{"", "5m-restore"}, wantError: "Postgres name must not be empty"},
		{name: "empty explicit branch is not reprompted", positionals: []string{"orders", ""}, wantError: "new branch name must not be empty"},
		{name: "whitespace explicit branch is not reprompted", positionals: []string{"orders", " \t "}, wantError: "new branch name must not be empty"},
		{name: "extra positional rejected", positionals: []string{"orders", "5m-restore", "extra"}, wantError: "expected a Postgres and optional new branch"},
		{name: "team validation before omitted-name prompt", positionals: []string{"orders"}, omitTeam: true, wantError: "missing required team"},
		{name: "time required before omitted-name prompt", positionals: []string{"orders"}, omitTime: true, wantError: "--at is required"},
		{name: "invalid time before omitted-name prompt", positionals: []string{"orders"}, omitTime: true, extraFlags: []string{"--ago", "yesterday"}, wantError: "--ago must be a positive duration"},
		{
			name: "supplied nonempty name remains for API validation", positionals: []string{"orders", " invalid name "}, apiError: "invalid branch name",
			wantError: "invalid branch name", wantOperations: []string{"CreatePostgresBranch"},
		},
		{
			name: "API failure", positionals: []string{"orders", "5m-restore"}, apiError: "access denied",
			wantError: "access denied", wantOperations: []string{"CreatePostgresBranch"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
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
			var operations []string
			var created gql.CreatePostgresBranchInput
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Team  string                        `json:"team"`
						Input gql.CreatePostgresBranchInput `json:"input"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetTeamPostgreses":
					if request.Variables.Team != "my-team" {
						t.Errorf("incorrect discovery team: %q", request.Variables.Team)
					}
					response = `{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null}],"pageInfo":{"hasNextPage":false}}}}}`
				case "CreatePostgresBranch":
					created = request.Variables.Input
					response = fmt.Sprintf(`{"data":{"createPostgresBranch":{"postgresBranch":{"name":%q,"state":"PROGRESSING"}}}}`, tt.returnedName)
					if tt.apiError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.apiError)
					}
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(response)); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			t.Setenv("NAIS_TEAM", "")
			t.Setenv("NAIS_ENVIRONMENT", "")
			config := filepath.Join(t.TempDir(), "config.yaml")
			contents := tt.config
			if contents == "" {
				contents = "{}\n"
			}
			if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("branch-name-test", "Test branch creation", "test", naistrix.ApplicationWithWriter(&output))
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
			args := []string{"--config", config, "--no-colors", "postgres", "branch", "create"}
			args = append(args, tt.positionals...)
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			if !tt.resolveEnvironment {
				args = append(args, "-e", "dev-gcp")
			}
			if !tt.omitTime {
				args = append(args, "--ago", "5m")
			}
			args = append(args, "--from", "main")
			args = append(args, tt.extraFlags...)
			before := time.Now()
			err = app.Run(naistrix.RunWithArgs(args))
			after := time.Now()
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v; want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(operations, "CreatePostgresBranch") {
				if created.Postgres != tt.positionals[0] || created.Branch != tt.positionals[1] || created.TeamSlug != "my-team" || created.EnvironmentName != "dev-gcp" || created.SourceBranch != "main" {
					t.Errorf("incorrect create input: %+v", created)
				}
				if created.TargetTime.Before(before.Add(-5*time.Minute).Truncate(time.Second)) || created.TargetTime.After(after.Add(-5*time.Minute).Truncate(time.Second)) || created.TargetTime.Location() != time.UTC || created.TargetTime.Nanosecond() != 0 {
					t.Errorf("incorrect relative recovery time: %v", created.TargetTime)
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}

func TestBranchCreateCompletionWithOptionalName(t *testing.T) {
	for _, tt := range []struct {
		name, environment string
		positionals       []string
		from              bool
		want              string
		wantOperations    []string
	}{
		{name: "Postgres names", want: "orders", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "no existing branch suggestions for new destination", positionals: []string{"orders"}},
		{name: "no suggestions after destination", positionals: []string{"orders", "restored"}},
		{name: "source flag before Postgres", from: true},
		{name: "source flag with sole Postgres positional", positionals: []string{"orders"}, environment: "dev-gcp", from: true, want: "main", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "source flag with optional destination present", positionals: []string{"orders", "restored"}, environment: "dev-gcp", from: true, want: "main", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "source flag resolves sole environment", positionals: []string{"orders"}, from: true, want: "main", wantOperations: []string{"GetTeamPostgreses", "GetPostgresBranchStatus"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string                                       `json:"operationName"`
					Variables     struct{ Team, Environment, Postgres string } `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetTeamPostgreses":
					response = `{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}},"activeBranch":null}],"pageInfo":{"hasNextPage":false}}}}}`
				case "GetPostgresBranchStatus":
					if request.Variables.Team != "my-team" || request.Variables.Postgres != "orders" || request.Variables.Environment != "dev-gcp" {
						t.Errorf("incorrect source completion target: %+v", request.Variables)
					}
					response = `{"data":{"team":{"environment":{"postgres":{"activeBranch":{"name":"main"},"desiredActiveBranch":"main","branches":{"nodes":[{"name":"main","state":"AVAILABLE"}]}}}}}}`
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(response)); err != nil {
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
			app, global, err := naistrix.NewApplication("branch-completion-test", "Test branch completion", "test", naistrix.ApplicationWithWriter(&output))
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
			args := []string{"--config", config, "-t", "my-team"}
			if tt.environment != "" {
				args = append(args, "-e", tt.environment)
			}
			args = append(args, "__complete", "postgres", "branch", "create")
			args = append(args, tt.positionals...)
			if tt.from {
				args = append(args, "--from")
			}
			args = append(args, "")
			if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
				t.Fatal(err)
			}
			if tt.want != "" && !strings.Contains(output.String(), tt.want+"\n") {
				t.Errorf("missing completion %q: %s", tt.want, output.String())
			}
			if tt.want == "" && (strings.Contains(output.String(), "orders\n") || strings.Contains(output.String(), "main\n")) {
				t.Errorf("unexpected existing-name completion: %s", output.String())
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
