package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestValkeyMutationOptionalFields(t *testing.T) {
	t.Setenv("NAIS_API_LOCAL_HOST", "api.nais.localhost")
	t.Setenv("NAIS_API_LOCAL_EMAIL", "")
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	for _, update := range []bool{false, true} {
		for _, tc := range []struct {
			name        string
			version     gql.ValkeyMajorVersion
			persistence *bool
			policy      gql.ValkeyMaxMemoryPolicy
		}{
			{name: "omitted"},
			{name: "enabled", version: gql.ValkeyMajorVersionV81, persistence: new(false)},
			{name: "disabled", version: gql.ValkeyMajorVersionV91, persistence: new(true)},
			{name: "best-effort", version: "V9_0"},
			{name: "max-memory-policy", policy: gql.ValkeyMaxMemoryPolicyAllkeysLru},
			{name: "no-eviction", policy: gql.ValkeyMaxMemoryPolicyNoEviction},
		} {
			t.Run(fmt.Sprintf("update=%t/%s", update, tc.name), func(t *testing.T) {
				calls := 0
				http.DefaultTransport = testTransport(func(r *http.Request) (*http.Response, error) {
					calls++
					var req struct {
						Variables struct {
							Input map[string]json.RawMessage `json:"input"`
						} `json:"variables"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						return nil, err
					}
					input := req.Variables.Input
					gotPolicy, present := input["maxMemoryPolicy"]
					if tc.policy == "" {
						if present {
							t.Errorf("maxMemoryPolicy should be omitted, got %s", gotPolicy)
						}
					} else if string(gotPolicy) != fmt.Sprintf("%q", tc.policy) {
						t.Errorf("maxMemoryPolicy = %s, want %q", gotPolicy, tc.policy)
					}
					wantVersion := tc.version
					if !update && wantVersion == "" {
						wantVersion = gql.ValkeyMajorVersionV91
					}
					gotVersion, present := input["version"]
					if wantVersion == "" {
						if present {
							t.Errorf("version should be omitted, got %s", gotVersion)
						}
					} else if string(gotVersion) != fmt.Sprintf("%q", wantVersion) {
						t.Errorf("version = %s, want %q", gotVersion, wantVersion)
					}
					gotPersistence, present := input["persistenceDisabled"]
					if tc.persistence == nil {
						if present {
							t.Errorf("persistenceDisabled should be omitted, got %s", gotPersistence)
						}
					} else if string(gotPersistence) != fmt.Sprint(*tc.persistence) {
						t.Errorf("persistenceDisabled = %s, want %t", gotPersistence, *tc.persistence)
					}
					field := "createValkey"
					if update {
						field = "updateValkey"
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body: io.NopCloser(strings.NewReader(fmt.Sprintf(
							`{"data":{"%s":{"valkey":{"id":"id","name":"my-cache"}}}}`, field))),
					}, nil
				})
				data := &Valkey{
					Memory: gql.ValkeyMemoryGb4, Tier: gql.ValkeyTierSingleNode,
					Version: tc.version, PersistenceDisabled: tc.persistence,
					MaxMemoryPolicy: tc.policy,
				}
				meta := Metadata{Name: "my-cache", TeamSlug: "my-team", EnvironmentName: "dev"}
				var err error
				if update {
					_, err = Update(context.Background(), meta, data)
				} else {
					_, err = Create(context.Background(), meta, data)
				}
				if err != nil {
					t.Fatal(err)
				}
				if calls != 1 {
					t.Errorf("request count = %d, want 1", calls)
				}
			})
		}
	}
}
