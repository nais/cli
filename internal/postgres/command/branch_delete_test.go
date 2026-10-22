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

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestBranchDeletionWarnsBeforeConfirmation(t *testing.T) {
	const unused = `{"data":{"team":{"environment":{"postgres":{"branch":{"workloads":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}}}}`
	const used = `{"data":{"team":{"environment":{"postgres":{"branch":{"workloads":{"nodes":[{"__typename":"Application","name":"preview-app"},{"__typename":"Job","name":"preview-job"}],"pageInfo":{"hasNextPage":false}}}}}}}}`
	const firstPage = `{"data":{"team":{"environment":{"postgres":{"branch":{"workloads":{"nodes":[{"__typename":"Application","name":"preview-app"}],"pageInfo":{"hasNextPage":true,"endCursor":"next-page"}}}}}}}}`
	const lastPage = `{"data":{"team":{"environment":{"postgres":{"branch":{"workloads":{"nodes":[{"__typename":"Job","name":"preview-job"}],"pageInfo":{"hasNextPage":false}}}}}}}}`
	for _, tt := range []struct {
		name, wantError   string
		pages             []string
		yes, warn, remove bool
	}{
		{name: "unused branch", pages: []string{unused}, yes: true, remove: true},
		{name: "yes skips confirmation but not usage warning", pages: []string{used}, yes: true, warn: true, remove: true},
		{name: "warning precedes unavailable confirmation", pages: []string{used}, warn: true, wantError: "no interactive terminal available"},
		{name: "usage lookup failure prevents deletion even with yes", pages: []string{`{"errors":[{"message":"access denied"}]}`}, yes: true, wantError: "access denied"},
		{name: "all pages are shown", pages: []string{firstPage, lastPage}, yes: true, warn: true, remove: true},
		{name: "later page failure prevents deletion", pages: []string{firstPage, `{"errors":[{"message":"lookup failed"}]}`}, yes: true, wantError: "lookup failed"},
		{name: "broken pagination prevents deletion", pages: []string{strings.ReplaceAll(firstPage, `"next-page"`, `null`)}, yes: true, wantError: "no new cursor"},
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
			var output bytes.Buffer
			var operations []string
			page := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Team, Environment, Postgres, Branch string
						After                               *string
						Input                               gql.DeletePostgresBranchInput
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetPostgresBranchWorkloads":
					if request.Variables.Team != "my-team" || request.Variables.Environment != "dev" || request.Variables.Postgres != "orders" || request.Variables.Branch != "preview" {
						t.Errorf("incorrect usage target: %+v", request.Variables)
					}
					if page == 0 && request.Variables.After != nil || page > 0 && (request.Variables.After == nil || *request.Variables.After != "next-page") {
						t.Errorf("incorrect page cursor: %v", request.Variables.After)
					}
					if page >= len(tt.pages) {
						t.Errorf("unexpected workload page %d", page)
						response = `{"errors":[{"message":"unexpected page"}]}`
					} else {
						response = tt.pages[page]
					}
					page++
				case "DeletePostgresBranch":
					if !tt.remove {
						t.Error("deletion requested without a successful usage check and confirmation")
					}
					if tt.warn && (!strings.Contains(output.String(), "preview-app") || !strings.Contains(output.String(), "preview-job") || !strings.Contains(output.String(), "connections unavailable")) {
						t.Errorf("usage warning was not printed before deletion: %s", output.String())
					}
					want := gql.DeletePostgresBranchInput{Postgres: "orders", Branch: "preview", TeamSlug: "my-team", EnvironmentName: "dev"}
					if request.Variables.Input != want {
						t.Errorf("delete input = %+v, want %+v", request.Variables.Input, want)
					}
					response = `{"data":{"deletePostgresBranch":{"postgresBranchDeleted":true}}}`
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
					response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, request.OperationName)
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
			app, global, err := naistrix.NewApplication("branch-delete-test", "Test branch deletion", "test", naistrix.ApplicationWithWriter(&output))
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
			args := []string{"--config", config, "--no-colors", "postgres", "branch", "delete", "orders", "preview", "-t", "my-team", "-e", "dev"}
			if tt.yes {
				args = append(args, "--yes")
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(output.String(), "connections unavailable") != tt.warn {
				t.Errorf("usage warning does not match expected usage: %s", output.String())
			}
			if slices.Contains(operations, "DeletePostgresBranch") != tt.remove {
				t.Errorf("unexpected operations: %v", operations)
			}
			if strings.Contains(output.String(), "Started deleting branch") != tt.remove {
				t.Errorf("incorrect deletion progress output: %s", output.String())
			}
		})
	}
}
