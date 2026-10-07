package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestBranchCreateSource(t *testing.T) {
	for _, tt := range []struct {
		name, from, statusResponse, wantSource, wantError string
		ago                                               string
		omitTime                                          bool
		wantOperations                                    []string
	}{
		{
			name:           "defaults to observed active branch during pending activation",
			statusResponse: `{"data":{"team":{"environment":{"postgres":{"activeBranch":{"name":"current"},"desiredActiveBranch":"pending","branches":{"nodes":[]}}}}}}`,
			wantSource:     "current", wantOperations: []string{"GetPostgresBranchStatus", "CreatePostgresBranch"},
		},
		{
			name: "explicit source overrides active branch without fetching status",
			from: "older", wantSource: "older", wantOperations: []string{"CreatePostgresBranch"},
		},
		{
			name:           "no active branch does not fall back to requested branch or main",
			statusResponse: `{"data":{"team":{"environment":{"postgres":{"activeBranch":null,"desiredActiveBranch":"main","branches":{"nodes":[]}}}}}}`,
			wantError:      "specify a source branch with --from", wantOperations: []string{"GetPostgresBranchStatus"},
		},
		{
			name:           "status lookup error prevents creation",
			statusResponse: `{"errors":[{"message":"access denied"}]}`,
			wantError:      "access denied", wantOperations: []string{"GetPostgresBranchStatus"},
		},
		{
			name:     "recovery time is still required before any API call",
			omitTime: true, wantError: "--at is required",
		},
		{
			name: "relative time is passed to the API as a UTC timestamp",
			ago:  "2h", omitTime: true,
			statusResponse: `{"data":{"team":{"environment":{"postgres":{"activeBranch":{"name":"current"},"desiredActiveBranch":"pending","branches":{"nodes":[]}}}}}}`,
			wantSource:     "current", wantOperations: []string{"GetPostgresBranchStatus", "CreatePostgresBranch"},
		},
		{
			name: "absolute and relative time cannot be combined",
			ago:  "2h", wantError: "--at and --ago cannot be used together",
		},
		{
			name: "invalid duration prevents API calls",
			ago:  "yesterday", omitTime: true, wantError: "--ago must be a positive duration",
		},
		{
			name: "negative duration prevents API calls",
			ago:  "-2h", omitTime: true, wantError: "--ago must be a positive duration",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			var created gql.CreatePostgresBranchInput
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Input gql.CreatePostgresBranchInput `json:"input"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				response := tt.statusResponse
				switch request.OperationName {
				case "GetPostgresBranchStatus":
				case "CreatePostgresBranch":
					created = request.Variables.Input
					response = `{"data":{"createPostgresBranch":{"postgresBranch":{"name":"restored","state":"PROGRESSING"}}}}`
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
			app, global, err := naistrix.NewApplication("branch-create-test", "Test branch creation", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			parent := &flag.Postgres{GlobalFlags: &flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional}}
			if err := app.AddCommand(branchCreateCommand(parent)); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "create", "orders", "restored", "-t", "my-team", "-e", "dev-gcp"}
			if !tt.omitTime {
				args = append(args, "--at", "2026-01-01T12:00:00Z")
			}
			if tt.from != "" {
				args = append(args, "--from", tt.from)
			}
			if tt.ago != "" {
				args = append(args, "--ago", tt.ago)
			}
			before := time.Now()
			err = app.Run(naistrix.RunWithArgs(args))
			after := time.Now()
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v; want %q", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if created.SourceBranch != tt.wantSource || created.Postgres != "orders" || created.Branch != "restored" || created.EnvironmentName != "dev-gcp" || created.TeamSlug != "my-team" {
					t.Errorf("unexpected create input: %+v", created)
				}
				if tt.ago != "" {
					if created.TargetTime.Before(before.Add(-2*time.Hour).Truncate(time.Second)) || created.TargetTime.After(after.Add(-2*time.Hour).Truncate(time.Second)) || created.TargetTime.Nanosecond() != 0 || created.TargetTime.Location() != time.UTC {
						t.Errorf("unexpected relative recovery time: %v", created.TargetTime)
					}
				}
				if !strings.Contains(output.String(), `Started creating branch "restored".`) {
					t.Errorf("missing creation result: %s", output.String())
				}
			}
			if !reflect.DeepEqual(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
