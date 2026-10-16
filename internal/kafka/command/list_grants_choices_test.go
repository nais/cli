package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestListGrantsDynamicTopicChoices(t *testing.T) {
	for _, tt := range []struct {
		name, wantError                string
		completion, completeAfterValue bool
		empty, apiError, unknown       bool
		omitTeam, omitEnvironment      bool
	}{
		{name: "existing topic"},
		{name: "unknown topic rejected before execution", unknown: true, wantError: "invalid value"},
		{name: "no available topics", empty: true, wantError: "no choices available"},
		{name: "API error preserved", apiError: true, wantError: "access denied"},
		{name: "missing team stops lookup", omitTeam: true, wantError: "missing required team"},
		{name: "missing environment stops lookup", omitEnvironment: true, wantError: "missing required environment"},
		{name: "completion offers existing topic", completion: true},
		{name: "completion filters by prefix", completion: true, unknown: true},
		{name: "completion reports no topics", completion: true, empty: true, wantError: "no choices available"},
		{name: "completion reports API error", completion: true, apiError: true, wantError: "access denied"},
		{name: "completion validates team before lookup", completion: true, omitTeam: true, wantError: "missing required team"},
		{name: "completion validates environment before lookup", completion: true, omitEnvironment: true, wantError: "missing required environment"},
		{name: "no lookup after final argument", completion: true, completeAfterValue: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
					Variables     struct {
						Team, TeamSlug, TopicName, EnvironmentName string
						Filter                                     struct{ Environments []string }
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				v := request.Variables
				var response string
				switch request.OperationName {
				case "GetTeamKafkaTopics":
					if v.Team != "my-team" || !slices.Equal(v.Filter.Environments, []string{"dev"}) {
						t.Errorf("incorrect topic lookup: %+v", v)
					}
					response = `{"data":{"team":{"kafkaTopics":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev"}}}]}}}}`
					if tt.empty {
						response = `{"data":{"team":{"kafkaTopics":{"nodes":[]}}}}`
					}
					if tt.apiError {
						response = `{"errors":[{"message":"access denied"}]}`
					}
				case "GetKafkaTopicGrants":
					if v.TeamSlug != "my-team" || v.TopicName != "orders" || v.EnvironmentName != "dev" {
						t.Errorf("incorrect grants target: %+v", v)
					}
					response = `{"data":{"team":{"kafkaTopics":{"nodes":[{"acl":{"nodes":[]}}]}}}}`
				default:
					t.Errorf("unexpected API operation: %s", request.OperationName)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, response); err != nil {
					t.Errorf("writing response: %v", err)
				}
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test Kafka topic choices", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Kafka(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"kafka", "list-grants"}
			if !tt.omitTeam {
				args = append(args, "-t", "my-team")
			}
			if !tt.omitEnvironment {
				args = append(args, "-e", "dev")
			}
			value := "orders"
			if tt.unknown {
				value = "unknown"
			}
			if tt.completion {
				args = append([]string{"__complete"}, args...)
				if tt.completeAfterValue {
					args = append(args, value)
				}
				prefix := ""
				if tt.unknown {
					prefix = value
				}
				args = append(args, prefix)
			} else {
				args = append(args, value)
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.completion {
				if err != nil {
					t.Fatal(err)
				}
				got := output.String()
				if tt.wantError != "" {
					if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) || strings.Count(got, "\n") != 2 {
						t.Errorf("completion = %q, want a single active help error containing %q", got, tt.wantError)
					}
				} else if tt.completeAfterValue {
					if got != ":4\n" {
						t.Errorf("completion = %q, want no suggestions", got)
					}
				} else {
					if strings.Contains(got, "orders\n") == tt.unknown {
						t.Errorf("completion = %q, unexpected suggestions for prefix", got)
					}
					if !strings.Contains(got, "_activeHelp_ Name of the Kafka topic to list grants for\n") {
						t.Errorf("completion = %q, want argument prompt", got)
					}
				}
				if !strings.HasSuffix(got, ":4\n") {
					t.Errorf("completion = %q, want NoFileComp", got)
				}
			} else if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			var wantOperations []string
			if !tt.omitTeam && !tt.omitEnvironment && !tt.completeAfterValue {
				wantOperations = append(wantOperations, "GetTeamKafkaTopics")
				if !tt.completion && tt.wantError == "" {
					wantOperations = append(wantOperations, "GetKafkaTopicGrants")
				}
			}
			if !slices.Equal(operations, wantOperations) {
				t.Errorf("API operations = %v, want %v", operations, wantOperations)
			}
		})
	}
}
