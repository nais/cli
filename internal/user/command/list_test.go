package command

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/user"
	"github.com/nais/naistrix"
)

func TestListUsers(t *testing.T) {
	firstPage := `{"data":{"users":{"nodes":[{"name":"Alice","email":"alice@example.com","source":"TENANT","isAdmin":true}],"pageInfo":{"hasNextPage":true,"endCursor":"next"}}}}`
	lastPage := `{"data":{"users":{"nodes":[{"name":"Bob","email":"bob@example.com","source":"NAIS","isAdmin":false}],"pageInfo":{"hasNextPage":false,"endCursor":"last"}}}}`
	unknownSourcePage := `{"data":{"users":{"nodes":[{"name":"Carol","email":"carol@example.com","source":null,"isAdmin":false}],"pageInfo":{"hasNextPage":false}}}}`
	emptyPage := `{"data":{"users":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`
	apiError := `{"errors":[{"message":"access denied"}]}`
	alice := user.User{Name: "Alice", Email: "alice@example.com", Source: new(gql.UserSourceTenant), IsAdmin: true}
	bob := user.User{Name: "Bob", Email: "bob@example.com", Source: new(gql.UserSourceNais)}
	for _, tt := range []struct {
		name        string
		command     string
		format      string
		args        []string
		pages       []string
		wantSources []string
		wantAdmin   string
		want        []user.User
		wantErr     string
		wantUsage   bool
	}{
		{
			name: "paginated table", command: "users", format: "table", pages: []string{firstPage, lastPage},
			want: []user.User{alice, bob},
		},
		{
			name: "paginated json", command: "users", format: "json", pages: []string{firstPage, lastPage},
			want: []user.User{alice, bob},
		},
		{
			name: "unrecorded source", command: "users", format: "json", pages: []string{unknownSourcePage},
			want: []user.User{{Name: "Carol", Email: "carol@example.com"}},
		},
		{name: "empty table", command: "users", format: "table", pages: []string{emptyPage}},
		{name: "empty json", command: "users", format: "json", pages: []string{emptyPage}, want: []user.User{}},
		{
			name: "one source", command: "users", format: "json", args: []string{"--source", "NAIS"},
			pages: []string{lastPage}, wantSources: []string{"NAIS"}, want: []user.User{bob},
		},
		{
			name: "repeated source across pages", command: "users", format: "json", args: []string{"--source", "tenant", "--source", "NAIS"},
			pages: []string{firstPage, lastPage}, wantSources: []string{"TENANT", "NAIS"},
			want: []user.User{alice, bob},
		},
		{
			name: "invalid source", command: "users", format: "json", args: []string{"--source", "other"},
			wantErr: `invalid --source "other"`,
		},
		{
			name: "admins only", command: "users", format: "json", args: []string{"--admin"},
			pages: []string{emptyPage}, wantAdmin: "true", want: []user.User{},
		},
		{
			name: "non-admins by source", command: "users", format: "json", args: []string{"--admin=false", "--source", "NAIS"},
			pages: []string{lastPage}, wantAdmin: "false", wantSources: []string{"NAIS"}, want: []user.User{bob},
		},
		{
			name: "invalid admin filter", command: "users", format: "json", args: []string{"--admin=other"},
			wantErr: `invalid argument "other"`, wantUsage: true,
		},
		{
			name: "singular alias", command: "user", format: "table", pages: []string{lastPage},
			want: []user.User{bob},
		},
		{name: "api error", command: "users", format: "table", pages: []string{apiError}, wantErr: "access denied"},
		{name: "later page error", command: "users", format: "json", pages: []string{firstPage, apiError}, wantErr: "access denied"},
		{
			name: "missing cursor", command: "users", format: "json",
			pages: []string{`{"data":{"users":{"nodes":[],"pageInfo":{"hasNextPage":true}}}}`}, wantErr: "no new cursor",
		},
		{name: "repeated cursor", command: "users", format: "json", pages: []string{firstPage, firstPage}, wantErr: "no new cursor"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query     string `json:"query"`
					Variables struct {
						After  *string `json:"after"`
						Filter *struct {
							Sources []string        `json:"sources"`
							IsAdmin json.RawMessage `json:"isAdmin"`
						} `json:"filter"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if !strings.Contains(request.Query, "users(first: 1000,") {
					t.Errorf("expected a page size of 1000 in query: %s", request.Query)
				}
				if requests == 0 && request.Variables.After != nil {
					t.Error("first request must not have a cursor")
				}
				if requests > 0 && (request.Variables.After == nil || *request.Variables.After != "next") {
					t.Error("subsequent request must use the previous page's cursor")
				}
				filter := request.Variables.Filter
				if tt.wantSources == nil && tt.wantAdmin == "" && filter != nil {
					t.Errorf("unexpected filter: %+v", filter)
				}
				if tt.wantSources != nil || tt.wantAdmin != "" {
					if filter == nil || !reflect.DeepEqual(filter.Sources, tt.wantSources) || string(filter.IsAdmin) != tt.wantAdmin {
						t.Errorf("filter = %+v, want sources %v and isAdmin %q", filter, tt.wantSources, tt.wantAdmin)
					}
				}
				if requests >= len(tt.pages) {
					t.Error("unexpected extra request")
					http.Error(w, "unexpected extra request", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.pages[requests]))
				requests++
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("NAIS_API_LOCAL_EMAIL", "tester@example.com")
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("users-test", "Test users", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Users(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", config, "--no-colors", tt.command, "list", "--output", tt.format}, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				if tt.wantUsage {
					if !strings.HasPrefix(output.String(), "Usage:") {
						t.Errorf("expected usage after a flag parse error: %s", output.String())
					}
				} else if output.Len() != 0 {
					t.Errorf("must not render partial results on error: %s", output.String())
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if tt.format == "json" {
					var got []user.User
					if err := json.Unmarshal(output.Bytes(), &got); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, tt.want) {
						t.Fatalf("unexpected users: %s", output.String())
					}
					if tt.name == "unrecorded source" && !strings.Contains(output.String(), `"source": null`) {
						t.Errorf("unrecorded source must render as null: %s", output.String())
					}
				} else if len(tt.want) == 0 {
					if !strings.Contains(output.String(), "No users found.") {
						t.Errorf("missing empty result message: %s", output.String())
					}
				} else {
					admins := 0
					for _, u := range tt.want {
						if u.IsAdmin {
							admins++
						}
						for _, value := range []string{u.Name, u.Email, string(*u.Source)} {
							if !strings.Contains(output.String(), value) {
								t.Errorf("table missing %q: %s", value, output.String())
							}
						}
					}
					if got := strings.Count(output.String(), "✓"); got != admins {
						t.Errorf("table has %d admin checkmarks, want %d: %s", got, admins, output.String())
					}
					if cells := strings.Fields(output.String()); slices.Contains(cells, "true") || slices.Contains(cells, "false") {
						t.Errorf("table shows raw admin booleans: %s", output.String())
					}
				}
			}
			if requests != len(tt.pages) {
				t.Errorf("made %d requests, want %d", requests, len(tt.pages))
			}
		})
	}
}
