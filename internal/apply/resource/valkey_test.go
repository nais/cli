package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValkeyApplyVersion(t *testing.T) {
	for _, exists := range []bool{false, true} {
		for _, tc := range []struct {
			version string
			want    string
		}{
			{version: "v8.1", want: "V8_1"},
			{version: "v9.1", want: "V9_1"},
			{version: "V8_1", want: "V8_1"},
			{version: "V9_1", want: "V9_1"},
			{},
		} {
			t.Run(fmt.Sprintf("exists=%t/version=%s", exists, tc.version), func(t *testing.T) {
				mutations := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var req struct {
						OperationName string `json:"operationName"`
						Variables     struct {
							Input map[string]json.RawMessage `json:"input"`
						} `json:"variables"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Errorf("decode request: %v", err)
						http.Error(w, "invalid request", http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					if req.OperationName == "GetValkey" {
						if exists {
							_, _ = w.Write([]byte(`{"data":{"team":{"environment":{"valkey":{"name":"my-cache"}}}}}`))
						} else {
							_, _ = w.Write([]byte(`{"errors":[{"message":"Resource not found: Valkey"}]}`))
						}
						return
					}

					mutations++
					operation, field := "CreateValkey", "createValkey"
					want := tc.want
					if exists {
						operation, field = "UpdateValkey", "updateValkey"
					} else if want == "" {
						want = "V9_1"
					}
					if req.OperationName != operation {
						t.Errorf("operation = %q, want %q", req.OperationName, operation)
					}
					got, present := req.Variables.Input["version"]
					if want == "" {
						if present {
							t.Errorf("version should be omitted on update, got %s", got)
						}
					} else if string(got) != fmt.Sprintf("%q", want) {
						t.Errorf("version = %s, want %q", got, want)
					}
					_, _ = fmt.Fprintf(w, `{"data":{"%s":{"valkey":{"id":"id","name":"my-cache"}}}}`, field)
				}))
				defer server.Close()
				t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))

				specVersion := ""
				if tc.version != "" {
					specVersion = "\n  version: " + tc.version
				}
				docs, err := Documents([]byte("version: v1\ntype: Valkey\nname: my-cache\nspec:\n  memory: 4GB\n  tier: SingleNode" + specVersion))
				if err != nil {
					t.Fatal(err)
				}
				m, err := ParseManifest(docs[0])
				if err != nil {
					t.Fatal(err)
				}
				action, err := (valkeyResource{}).Apply(context.Background(), Metadata{
					TeamSlug: "my-team", EnvironmentName: "dev",
				}, m)
				if err != nil {
					t.Fatal(err)
				}
				wantAction := ActionCreated
				if exists {
					wantAction = ActionUpdated
				}
				if action != wantAction {
					t.Errorf("action = %q, want %q", action, wantAction)
				}
				if mutations != 1 {
					t.Errorf("mutation count = %d, want 1", mutations)
				}
			})
		}
	}
}

func TestValkeyApplyRejectsInvalidVersion(t *testing.T) {
	docs, err := Documents([]byte("version: v1\ntype: Valkey\nname: my-cache\nspec:\n  memory: 4GB\n  tier: SingleNode\n  version: v10.1"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := ParseManifest(docs[0])
	if err != nil {
		t.Fatal(err)
	}
	_, err = (valkeyResource{}).Apply(context.Background(), Metadata{}, m)
	if err == nil || !strings.Contains(err.Error(), `invalid version "v10.1"`) {
		t.Fatalf("expected invalid version error, got %v", err)
	}
}
