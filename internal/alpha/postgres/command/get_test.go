package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestPostgresGetCommand(t *testing.T) {
	const configured = `{"name":"orders","majorVersion":"18","highAvailability":true,"resources":{"cpu":"200m","memory":"1Gi","diskSize":"20Gi"},"labels":[{"key":"domain","value":"payments"}],"activeBranch":{"name":"main"},"desiredActiveBranch":"restored","branches":{"nodes":[{"name":"main","state":"AVAILABLE"},{"name":"restored","state":"PROGRESSING"}]}}`
	const unconfigured = `{"name":"orders","majorVersion":"18","highAvailability":false,"resources":{"cpu":null,"memory":null,"diskSize":null},"labels":[],"activeBranch":null,"desiredActiveBranch":null,"branches":{"nodes":[]}}`
	const activated = `{"name":"orders","majorVersion":"18","highAvailability":false,"resources":{},"labels":[],"activeBranch":{"name":"main"},"desiredActiveBranch":"main","branches":{"nodes":[{"name":"main","state":"AVAILABLE"}]}}`
	readyWithFailedInactive := strings.Replace(activated, `{"name":"main","state":"AVAILABLE"}`, `{"name":"main","state":"AVAILABLE"},{"name":"failed-restore","state":"DEGRADED"}`, 1)
	pendingHealthy := strings.Replace(configured, `"state":"PROGRESSING"`, `"state":"AVAILABLE"`, 1)
	pendingDegraded := strings.Replace(configured, `"state":"PROGRESSING"`, `"state":"DEGRADED"`, 1)
	activeDegraded := strings.Replace(pendingHealthy, `"name":"main","state":"AVAILABLE"`, `"name":"main","state":"DEGRADED"`, 1)
	initial := strings.Replace(activated, `"activeBranch":{"name":"main"}`, `"activeBranch":null`, 1)
	initial = strings.Replace(initial, `"state":"AVAILABLE"`, `"state":"PROGRESSING"`, 1)
	missingActive := strings.Replace(activated, `{"name":"main","state":"AVAILABLE"}`, "", 1)
	futureState := strings.Replace(activated, `"state":"AVAILABLE"`, `"state":"FUTURE_STATE"`, 1)
	for _, tt := range []struct {
		name, environment, config, format, metadata, lookupError, getError, wantError, wantStatus string
		lookupEnvironments                                                                        []string
		complete, omitTeam, omitName                                                              bool
		wantOperations                                                                            []string
	}{
		{name: "configured metadata as text", environment: "dev-gcp", metadata: configured, wantStatus: "Updating", wantOperations: []string{"GetPostgres"}},
		{name: "configured metadata as JSON", environment: "dev-gcp", format: "json", metadata: configured, wantStatus: "Updating", wantOperations: []string{"GetPostgres"}},
		{name: "unset requests are not inferred in text", environment: "dev-gcp", metadata: unconfigured, wantStatus: "Unknown", wantOperations: []string{"GetPostgres"}},
		{name: "unset requests remain null in JSON", environment: "dev-gcp", format: "json", metadata: unconfigured, wantStatus: "Unknown", wantOperations: []string{"GetPostgres"}},
		{name: "observed activation omits requested branch", environment: "dev-gcp", format: "json", metadata: activated, wantStatus: "Ready", wantOperations: []string{"GetPostgres"}},
		{name: "observed activation text", environment: "dev-gcp", metadata: activated, wantStatus: "Ready", wantOperations: []string{"GetPostgres"}},
		{name: "inactive degraded branch does not degrade active summary", environment: "dev-gcp", metadata: readyWithFailedInactive, wantStatus: "Ready", wantOperations: []string{"GetPostgres"}},
		{name: "inactive degraded branch remains in JSON", environment: "dev-gcp", format: "json", metadata: readyWithFailedInactive, wantStatus: "Ready", wantOperations: []string{"GetPostgres"}},
		{name: "healthy requested branch still awaits activation", environment: "dev-gcp", metadata: pendingHealthy, wantStatus: "Updating", wantOperations: []string{"GetPostgres"}},
		{name: "pending degraded branch JSON", environment: "dev-gcp", format: "json", metadata: pendingDegraded, wantStatus: "Needs attention", wantOperations: []string{"GetPostgres"}},
		{name: "active degradation remains visible during pending activation", environment: "dev-gcp", metadata: activeDegraded, wantStatus: "Needs attention", wantOperations: []string{"GetPostgres"}},
		{name: "initial provisioning is not updating", environment: "dev-gcp", metadata: initial, wantStatus: "Not ready yet", wantOperations: []string{"GetPostgres"}},
		{name: "missing active state JSON", environment: "dev-gcp", format: "json", metadata: missingActive, wantStatus: "Unknown", wantOperations: []string{"GetPostgres"}},
		{name: "future state is unknown in JSON", environment: "dev-gcp", format: "json", metadata: futureState, wantStatus: "Unknown", wantOperations: []string{"GetPostgres"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", format: "json", metadata: activated, wantStatus: "Ready", wantOperations: []string{"GetPostgres"}},
		{name: "explicit environment overrides default", environment: "dev-gcp", config: "environment: prod-gcp\n", format: "json", metadata: configured, wantStatus: "Updating", wantOperations: []string{"GetPostgres"}},
		{name: "environment resolved from existing Postgres", lookupEnvironments: []string{"dev-gcp"}, format: "json", metadata: activated, wantStatus: "Ready", wantOperations: []string{"GetTeamPostgreses", "GetPostgres"}},
		{name: "ambiguous environment requires choice", lookupEnvironments: []string{"dev-gcp", "prod-gcp"}, wantError: "specify environment with -e, --environment", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "missing Postgres", wantError: `Postgres "orders" not found`, wantOperations: []string{"GetTeamPostgreses"}},
		{name: "discovery error", lookupError: "access denied", wantError: "access denied", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "get error", environment: "dev-gcp", getError: "Resource not found", wantError: "Resource not found", wantOperations: []string{"GetPostgres"}},
		{name: "team is required", omitTeam: true, wantError: "missing required team"},
		{name: "name is required before API calls", omitName: true, wantError: "Expected exactly 1 argument"},
		{name: "registered completion includes branchless Postgres", complete: true, lookupEnvironments: []string{"dev-gcp"}, wantOperations: []string{"GetTeamPostgreses"}},
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string                                       `json:"operationName"`
					Query         string                                       `json:"query"`
					Variables     struct{ Team, Environment, Postgres string } `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				var response string
				switch request.OperationName {
				case "GetTeamPostgreses":
					if request.Variables.Team != "my-team" {
						t.Errorf("incorrect discovery team: %+v", request.Variables)
					}
					nodes := make([]string, 0, len(tt.lookupEnvironments))
					for _, env := range tt.lookupEnvironments {
						nodes = append(nodes, fmt.Sprintf(`{"name":"orders","teamEnvironment":{"environment":{"name":%q}},"activeBranch":null}`, env))
					}
					response = `{"data":{"team":{"postgreses":{"nodes":[` + strings.Join(nodes, ",") + `],"pageInfo":{"hasNextPage":false}}}}}`
					if tt.lookupError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.lookupError)
					}
				case "GetPostgres":
					if request.Variables.Team != "my-team" || request.Variables.Environment != "dev-gcp" || request.Variables.Postgres != "orders" {
						t.Errorf("incorrect metadata target: %+v", request.Variables)
					}
					for _, sensitive := range []string{"password", "relayToken", "postgresAccess"} {
						if strings.Contains(request.Query, sensitive) {
							t.Errorf("metadata query requests sensitive access field %s", sensitive)
						}
					}
					response = `{"data":{"team":{"environment":{"postgres":` + tt.metadata + `}}}}`
					if tt.getError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.getError)
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
			config := filepath.Join(t.TempDir(), "config.yaml")
			contents := tt.config
			if contents == "" {
				contents = "{}\n"
			}
			if err := os.WriteFile(config, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("postgres-get-test", "Test Postgres metadata", "test", naistrix.ApplicationWithWriter(&output))
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
			args := []string{"--config", config, "--no-colors"}
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			if tt.complete {
				args = append(args, "__complete", "postgres", "get", "")
			} else {
				args = append(args, "postgres", "get")
				if !tt.omitName {
					args = append(args, "orders")
				}
				if tt.environment != "" {
					args = append(args, "-e", tt.environment)
				}
				if tt.format != "" {
					args = append(args, "--output", tt.format)
				}
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v; want %q", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				var expected struct {
					Name, MajorVersion  string
					HighAvailability    bool
					Resources           postgresResourceRequests
					Labels              []postgresLabel
					ActiveBranch        *struct{ Name string }
					DesiredActiveBranch *string
					Branches            struct{ Nodes []postgres.Branch }
				}
				if !tt.complete {
					if err := json.Unmarshal([]byte(tt.metadata), &expected); err != nil {
						t.Fatal(err)
					}
				}
				switch {
				case tt.complete:
					if !strings.Contains(output.String(), "orders\n") {
						t.Errorf("missing name completion: %s", output.String())
					}
				case tt.format == "json":
					var detail postgresDetails
					if err := json.Unmarshal(output.Bytes(), &detail); err != nil {
						t.Fatalf("invalid JSON %q: %v", output.String(), err)
					}
					if detail.Name != expected.Name || detail.Team != "my-team" || detail.Environment != "dev-gcp" || detail.Version != expected.MajorVersion || detail.Status != tt.wantStatus || detail.HighAvailability != expected.HighAvailability || !reflect.DeepEqual(detail.Resources, expected.Resources) || !slices.Equal(detail.Labels, expected.Labels) {
						t.Errorf("incorrect metadata/status: %+v", detail)
					}
					active, requested := "", ""
					if expected.ActiveBranch != nil {
						active = expected.ActiveBranch.Name
					}
					if expected.DesiredActiveBranch != nil && *expected.DesiredActiveBranch != active {
						requested = *expected.DesiredActiveBranch
					}
					if detail.ActiveBranch != active || detail.RequestedBranch != requested || len(detail.Branches) != len(expected.Branches.Nodes) {
						t.Fatalf("incorrect branch observations: %+v", detail)
					}
					for i, branch := range detail.Branches {
						if branch.Branch != expected.Branches.Nodes[i] || branch.Active != (branch.Name == active) || branch.Requested != (branch.Name == requested) {
							t.Errorf("incorrect branch or markers: %+v", branch)
						}
					}
					var fields map[string]json.RawMessage
					if err := json.Unmarshal(output.Bytes(), &fields); err != nil {
						t.Fatal(err)
					}
					for _, key := range []string{"name", "team", "environment", "version", "status", "highAvailability", "resources", "labels", "activeBranch", "branches"} {
						if _, ok := fields[key]; !ok {
							t.Errorf("missing JSON field %q: %s", key, output.String())
						}
					}
					if string(fields["labels"]) == "null" || string(fields["branches"]) == "null" {
						t.Errorf("JSON collections must remain arrays: %s", output.String())
					}
					if requested == "" && fields["requestedBranch"] != nil {
						t.Errorf("unexpected requested activation: %s", output.String())
					}
				default:
					for _, field := range []string{"orders · dev-gcp", "Status:", "Version:", "18", "High availability:", "CPU:", "Memory:", "Disk:"} {
						if !strings.Contains(output.String(), field) {
							t.Errorf("missing detail %s: %s", field, output.String())
						}
					}
					if !strings.Contains(output.String(), fmt.Sprintf("  %-18s %s\n", "Status:", tt.wantStatus)) {
						t.Errorf("missing readiness %q: %s", tt.wantStatus, output.String())
					}
					ha := "No"
					if expected.HighAvailability {
						ha = "Yes"
					}
					if !strings.Contains(output.String(), fmt.Sprintf("  %-18s %s\n", "High availability:", ha)) {
						t.Errorf("incorrect HA display: %s", output.String())
					}
					for _, resource := range []struct {
						label string
						value *string
					}{
						{"CPU:", expected.Resources.CPU}, {"Memory:", expected.Resources.Memory}, {"Disk:", expected.Resources.DiskSize},
					} {
						value := "(not configured)"
						if resource.value != nil {
							value = *resource.value
						}
						if !strings.Contains(output.String(), fmt.Sprintf("  %-18s %s\n", resource.label, value)) {
							t.Errorf("incorrect resource display for %s: %s", resource.label, output.String())
						}
					}
					if strings.Contains(output.String(), "Labels:") != (len(expected.Labels) > 0) {
						t.Errorf("incorrect labels visibility: %s", output.String())
					}
					for _, label := range expected.Labels {
						if !strings.Contains(output.String(), label.Key+"="+label.Value) {
							t.Errorf("missing compact label: %s", output.String())
						}
					}
					for _, branch := range expected.Branches.Nodes {
						if slices.Contains(strings.Fields(output.String()), branch.Name) {
							t.Errorf("human output includes branch name %q: %s", branch.Name, output.String())
						}
					}
					for _, forbidden := range []string{"Field", "Value", "Team", "my-team", "Active branch", "Requested branch", "Branches", "\nmain", "restored", "No labels configured", "No Postgres branches found", "derived", "SQL", "Reason:", "Scope", "readiness"} {
						if strings.Contains(output.String(), forbidden) {
							t.Errorf("human output includes unwanted detail %q: %s", forbidden, output.String())
						}
					}
				}
			}
			if strings.Contains(output.String(), "\x1b[") {
				t.Errorf("--no-colors emitted ANSI escapes: %q", output.String())
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}
