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

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestPostgresReadiness(t *testing.T) {
	available := gql.PostgresBranchStateAvailable
	progressing := gql.PostgresBranchStateProgressing
	degraded := gql.PostgresBranchStateDegraded
	for _, tt := range []struct {
		name, active, desired, wantStatus, wantReason string
		branches                                      []postgres.Branch
	}{
		{name: "available active branch", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Ready"},
		{name: "inactive degraded branch does not degrade active summary", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "failed-restore", State: degraded}}, wantStatus: "Ready"},
		{name: "progressing active branch", active: "main", branches: []postgres.Branch{{Name: "main", State: progressing}}, wantStatus: "Progressing", wantReason: `Active branch "main" is progressing.`},
		{name: "degraded active branch", active: "main", branches: []postgres.Branch{{Name: "main", State: degraded}}, wantStatus: "Degraded", wantReason: `Active branch "main" is degraded.`},
		{name: "available requested branch still awaits observed activation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: available}}, wantStatus: "Progressing", wantReason: `Activation of requested branch "restored" is pending.`},
		{name: "progressing requested branch", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: progressing}}, wantStatus: "Progressing", wantReason: `Activation of requested branch "restored" is pending.`},
		{name: "degraded requested branch", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}, {Name: "restored", State: degraded}}, wantStatus: "Degraded", wantReason: `Requested branch "restored" is degraded; activation is pending.`},
		{name: "pending activation does not hide active degradation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: degraded}, {Name: "restored", State: available}}, wantStatus: "Degraded", wantReason: `Active branch "main" is degraded.`},
		{name: "provisioning before active branch observation", desired: "main", branches: []postgres.Branch{{Name: "main", State: progressing}}, wantStatus: "Progressing", wantReason: `Activation of requested branch "main" is pending.`},
		{name: "available desired branch is not yet observed", desired: "main", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Progressing", wantReason: `Activation of requested branch "main" is pending.`},
		{name: "failed initial provisioning", desired: "main", branches: []postgres.Branch{{Name: "main", State: degraded}}, wantStatus: "Degraded", wantReason: `Requested branch "main" is degraded; activation is pending.`},
		{name: "requested branch not yet listed", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: available}}, wantStatus: "Progressing", wantReason: `Activation of requested branch "restored" is pending.`},
		{name: "no observed or desired branch", wantStatus: "Unknown", wantReason: "No observed active branch is known."},
		{name: "active branch missing from branch list", active: "main", branches: []postgres.Branch{{Name: "restored", State: available}}, wantStatus: "Unknown", wantReason: `Active branch "main" has no known state.`},
		{name: "active branch missing during pending activation", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "restored", State: available}}, wantStatus: "Unknown", wantReason: `Active branch "main" has no known state.`},
		{name: "active state missing", active: "main", branches: []postgres.Branch{{Name: "main"}}, wantStatus: "Unknown", wantReason: `Active branch "main" has no known state.`},
		{name: "unknown active state", active: "main", branches: []postgres.Branch{{Name: "main", State: "FUTURE_STATE"}}, wantStatus: "Unknown", wantReason: `Active branch "main" has no known state.`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state, reason := summarizePostgresStatus(postgres.BranchStatus{Active: tt.active, DesiredActive: tt.desired, Branches: tt.branches})
			if state != tt.wantStatus || reason != tt.wantReason {
				t.Errorf("status = %q, reason = %q; want %q, %q", state, reason, tt.wantStatus, tt.wantReason)
			}
		})
	}
}

func TestPostgresStatusCommand(t *testing.T) {
	for _, tt := range []struct {
		name, active, desired, format, wantStatus, wantError, statusError, discoveryError string
		environment, config                                                               string
		branches                                                                          []postgres.Branch
		lookupEnvironments                                                                []string
		omitTeam                                                                          bool
		complete                                                                          bool
		wantOperations                                                                    []string
	}{
		{name: "explicit environment ready text", environment: "dev-gcp", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: gql.PostgresBranchStateAvailable}, {Name: "failed-restore", State: gql.PostgresBranchStateDegraded}}, wantStatus: "Ready", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "pending activation text", environment: "dev-gcp", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: gql.PostgresBranchStateAvailable}, {Name: "restored", State: gql.PostgresBranchStateAvailable}}, wantStatus: "Progressing", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "pending degraded branch JSON", environment: "dev-gcp", format: "json", active: "main", desired: "restored", branches: []postgres.Branch{{Name: "main", State: gql.PostgresBranchStateAvailable}, {Name: "restored", State: gql.PostgresBranchStateDegraded}}, wantStatus: "Degraded", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "missing state JSON", environment: "dev-gcp", format: "json", active: "main", desired: "main", wantStatus: "Unknown", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "environment resolved from existing Postgres", lookupEnvironments: []string{"dev-gcp"}, format: "json", active: "main", desired: "main", branches: []postgres.Branch{{Name: "main", State: gql.PostgresBranchStateAvailable}}, wantStatus: "Ready", wantOperations: []string{"GetTeamPostgreses", "GetPostgresBranchStatus"}},
		{name: "default environment bypasses discovery", config: "environment: dev-gcp\n", format: "json", active: "main", branches: []postgres.Branch{{Name: "main", State: gql.PostgresBranchStateAvailable}}, wantStatus: "Ready", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "unknown without branches", environment: "dev-gcp", wantStatus: "Unknown", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "ambiguous environment", lookupEnvironments: []string{"dev-gcp", "prod-gcp"}, wantError: "specify environment with -e, --environment", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "missing Postgres", wantError: `Postgres "orders" not found`, wantOperations: []string{"GetTeamPostgreses"}},
		{name: "environment lookup failure", discoveryError: "access denied", wantError: "access denied", wantOperations: []string{"GetTeamPostgreses"}},
		{name: "status lookup failure", environment: "dev-gcp", statusError: "Resource not found", wantError: "Resource not found", wantOperations: []string{"GetPostgresBranchStatus"}},
		{name: "team is required", omitTeam: true, wantError: "missing required team"},
		{name: "registered name completion includes branchless Postgres", complete: true, lookupEnvironments: []string{"dev-gcp"}, wantOperations: []string{"GetTeamPostgreses"}},
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
					if tt.discoveryError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.discoveryError)
					}
				case "GetPostgresBranchStatus":
					if request.Variables.Team != "my-team" || request.Variables.Environment != "dev-gcp" || request.Variables.Postgres != "orders" {
						t.Errorf("incorrect status target: %+v", request.Variables)
					}
					for _, field := range []string{"majorVersion", "highAvailability", "resources"} {
						if strings.Contains(request.Query, field) {
							t.Errorf("status query includes settings field %s", field)
						}
					}
					nodes := make([]postgres.Branch, 0, len(tt.branches))
					nodes = append(nodes, tt.branches...)
					branches, err := json.Marshal(nodes)
					if err != nil {
						t.Errorf("encoding branches: %v", err)
					}
					active := "null"
					if tt.active != "" {
						active = fmt.Sprintf(`{"name":%q}`, tt.active)
					}
					response = fmt.Sprintf(`{"data":{"team":{"environment":{"postgres":{"activeBranch":%s,"desiredActiveBranch":%q,"branches":{"nodes":%s}}}}}}`, active, tt.desired, branches)
					if tt.statusError != "" {
						response = fmt.Sprintf(`{"errors":[{"message":%q}]}`, tt.statusError)
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
			app, global, err := naistrix.NewApplication("postgres-status-test", "Test Postgres status", "test", naistrix.ApplicationWithWriter(&output))
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
				args = append(args, "__complete", "postgres", "status", "")
			} else {
				args = append(args, "postgres", "status", "orders")
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
				switch {
				case tt.complete:
					if !strings.Contains(output.String(), "orders\n") {
						t.Errorf("missing Postgres name completion: %s", output.String())
					}
				case tt.format == "json":
					var result postgresStatusResult
					if err := json.Unmarshal(output.Bytes(), &result); err != nil {
						t.Fatalf("invalid JSON %q: %v", output.String(), err)
					}
					if result.Name != "orders" || result.Team != "my-team" || result.Environment != "dev-gcp" || result.Status != tt.wantStatus || result.ActiveBranch != tt.active {
						t.Errorf("incorrect status result: %+v", result)
					}
					requested := ""
					if tt.desired != "" && tt.desired != tt.active {
						requested = tt.desired
					}
					if result.RequestedBranch != requested || (result.Status != "Ready" && result.Reason == "") || result.Scope != postgresStatusScope || len(result.Branches) != len(tt.branches) {
						t.Fatalf("incorrect status detail: %+v", result)
					}
					for i, branch := range result.Branches {
						if branch.Branch != tt.branches[i] || branch.Active != (branch.Name == tt.active) || branch.Requested != (branch.Name == requested) {
							t.Errorf("incorrect branch or markers: %+v", branch)
						}
					}
					for _, field := range []string{`"version"`, `"highAvailability"`, `"resources"`} {
						if strings.Contains(output.String(), field) {
							t.Errorf("status includes settings: %s", output.String())
						}
					}
				default:
					if !strings.Contains(output.String(), "Status: "+tt.wantStatus) || !strings.Contains(output.String(), postgresStatusScope) || !strings.Contains(output.String(), "Active branch:") {
						t.Errorf("missing summary or scope: %s", output.String())
					}
					if tt.wantStatus != "Ready" && !strings.Contains(output.String(), "Reason:") {
						t.Errorf("missing non-ready reason: %s", output.String())
					}
					if tt.wantStatus == "Ready" && strings.Contains(output.String(), "Reason:") {
						t.Errorf("unexpected ready reason: %s", output.String())
					}
					pending := tt.desired != "" && tt.desired != tt.active
					if strings.Contains(output.String(), "Requested branch:") != pending {
						t.Errorf("incorrect requested branch visibility: %s", output.String())
					}
					if len(tt.branches) > 0 {
						for _, heading := range []string{"Name", "State", "Active", "Requested"} {
							if !strings.Contains(output.String(), heading) {
								t.Errorf("missing table heading %s: %s", heading, output.String())
							}
						}
						for _, branch := range tt.branches {
							if !strings.Contains(output.String(), branch.Name) || !strings.Contains(strings.ToLower(output.String()), strings.ToLower(string(branch.State))) {
								t.Errorf("missing branch in table: %s", output.String())
							}
						}
						markers := 0
						if tt.active != "" {
							markers++
						}
						if pending {
							markers++
						}
						if strings.Count(output.String(), "Yes") != markers {
							t.Errorf("incorrect active/requested markers: %s", output.String())
						}
					} else if !strings.Contains(output.String(), "No Postgres branches found.") {
						t.Errorf("missing empty branches result: %s", output.String())
					}
				}
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v; want %v", operations, tt.wantOperations)
			}
		})
	}
}

func TestPostgresStatusProgressCommand(t *testing.T) {
	for _, tt := range []struct{ name, team, environment, want string }{
		{"orders", "my-team", "dev-gcp", "nais alpha postgres status orders -t my-team -e dev-gcp"},
		{"orders", "my-team", "env with spaces", "nais alpha postgres status orders -t my-team -e 'env with spaces'"},
		{"orders", "my-team", "env'$(touch file)", "nais alpha postgres status orders -t my-team -e 'env'\"'\"'$(touch file)'"},
	} {
		if got := postgresStatusCommandLine(tt.name, tt.team, tt.environment); got != tt.want {
			t.Errorf("progress command = %q; want %q", got, tt.want)
		}
	}
}
