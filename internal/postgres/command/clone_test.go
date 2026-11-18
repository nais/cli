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
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestResolveCloneTime(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 16, 12, 0, 30, 500, time.UTC)
	for _, tt := range []struct {
		name, value, wantError string
		want                   time.Time
		explicit               bool
	}{
		{name: "default is ten minutes ago in whole seconds", want: time.Date(2026, 11, 16, 11, 50, 30, 0, time.UTC)},
		{name: "RFC 3339 UTC", value: "2026-11-16T10:00:00Z", want: time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC), explicit: true},
		{name: "RFC 3339 offset is normalized", value: "2026-11-16T11:00:00+01:00", want: time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC), explicit: true},
		{name: "local time without offset", value: "2026-11-16 11:00", want: time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC), explicit: true},
		{name: "local time with T and seconds", value: "2026-11-16T11:00:05", want: time.Date(2026, 11, 16, 10, 0, 5, 0, time.UTC), explicit: true},
		{name: "summer time uses the summer offset", value: "2026-07-01 12:00", want: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC), explicit: true},
		{name: "ambiguous autumn hour resolves to the earlier occurrence", value: "2026-10-25 02:30", want: time.Date(2026, 10, 25, 0, 30, 0, 0, time.UTC), explicit: true},
		{name: "ambiguous hour is fine with an explicit offset", value: "2026-10-25T02:30:00+01:00", want: time.Date(2026, 10, 25, 1, 30, 0, 0, time.UTC), explicit: true},
		{name: "hour after the autumn change is unambiguous", value: "2026-10-25 03:30", want: time.Date(2026, 10, 25, 2, 30, 0, 0, time.UTC), explicit: true},
		{name: "hour skipped by daylight saving is rejected", value: "2026-03-29 02:30", wantError: "does not exist", explicit: true},
		{name: "not a time", value: "yesterday", wantError: "--at must be a local time", explicit: true},
		{name: "date only", value: "2026-11-16", wantError: "--at must be a local time", explicit: true},
		{name: "fractional seconds", value: "2026-11-16T10:00:00.5Z", wantError: "whole seconds", explicit: true},
		{name: "future", value: "2026-11-16 13:01", wantError: "future", explicit: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, explicit, err := resolveCloneTime(tt.value, now, oslo)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil || !got.Equal(tt.want) || got.Location() != time.UTC || explicit != tt.explicit {
				t.Fatalf("got %v explicit=%t, %v; want %v explicit=%t in UTC", got, explicit, err, tt.want, tt.explicit)
			}
		})
	}
}

func TestValidateCloneWindow(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}
	earliest := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	if err := validateCloneWindow(earliest, true, &earliest, oslo); err != nil {
		t.Errorf("the earliest instant itself is cloneable: %v", err)
	}
	err = validateCloneWindow(earliest.Add(-time.Second), true, &earliest, oslo)
	if err == nil || !strings.Contains(err.Error(), "2026-11-10 01:00:00 CET") {
		t.Errorf("explicit before window must name the earliest time in local time: %v", err)
	}
	err = validateCloneWindow(earliest.Add(-time.Second), false, &earliest, oslo)
	if err == nil || !strings.Contains(err.Error(), "default clone point") {
		t.Errorf("default before window must explain it was the default: %v", err)
	}
	if err := validateCloneWindow(earliest, true, nil, oslo); err == nil || !strings.Contains(err.Error(), "no recoverable backup window") {
		t.Errorf("unknown window: %v", err)
	}
}

func TestCloneSettingExplainsInheritance(t *testing.T) {
	for _, tt := range []struct{ inherited, explicit, want string }{
		{"20Gi", "", "20Gi (from source)"},
		{"20Gi", "50Gi", "50Gi (overrides source)"},
		{"", "", "platform default"},
	} {
		if got := cloneSetting(tt.inherited, tt.explicit); got != tt.want {
			t.Errorf("cloneSetting(%q, %q) = %q, want %q", tt.inherited, tt.explicit, got, tt.want)
		}
	}
}

func TestCloneCommand(t *testing.T) {
	const source = `{"name":"orders","majorVersion":"17","state":"AVAILABLE","highAvailability":true,"earliestRestoreTime":"2026-11-10T00:00:00Z","resources":{"cpu":"200m","memory":"1Gi","diskSize":"20Gi"}}`
	for _, tt := range []struct {
		name       string
		target     string
		extra      []string
		source     string
		wantError  string
		wantCreate bool
		check      func(t *testing.T, input map[string]any)
	}{
		{
			name: "inherits everything and sends the source version", extra: []string{"--at", "2026-11-12T10:00:00Z"}, source: source, wantCreate: true,
			check: func(t *testing.T, input map[string]any) {
				recovery, _ := input["recovery"].(map[string]any)
				if recovery["sourcePostgres"] != "orders" || recovery["targetTime"] != "2026-11-12T10:00:00Z" {
					t.Errorf("recovery = %v", recovery)
				}
				if input["majorVersion"] != "17" || input["name"] != "orders-restored" || input["environmentName"] != "dev-gcp" || input["teamSlug"] != "my-team" {
					t.Errorf("input = %v", input)
				}
				for _, inherited := range []string{"highAvailability", "cpu", "memory", "diskSize"} {
					if _, set := input[inherited]; set {
						t.Errorf("%s must be omitted so the API inherits it from the source", inherited)
					}
				}
			},
		},
		{
			name: "explicit overrides are sent", extra: []string{"--at", "2026-11-12T10:00:00Z", "--cpu", "1", "--memory", "4Gi", "--disk-size", "50Gi", "--high-availability", "false"}, source: source, wantCreate: true,
			check: func(t *testing.T, input map[string]any) {
				if input["cpu"] != "1" || input["memory"] != "4Gi" || input["diskSize"] != "50Gi" || input["highAvailability"] != false {
					t.Errorf("overrides missing: %v", input)
				}
			},
		},
		{name: "time before the backup window never creates", extra: []string{"--at", "2026-11-09T23:59:59Z"}, source: source, wantError: "earliest possible clone time"},
		{name: "source without a backup window never creates", extra: []string{"--at", "2026-11-12T10:00:00Z"}, source: strings.Replace(source, `"2026-11-10T00:00:00Z"`, `null`, 1), wantError: "no recoverable backup window"},
		{
			name: "no time clones from the latest available point", source: strings.Replace(source, `"2026-11-10T00:00:00Z"`, `"2020-01-01T00:00:00Z"`, 1), wantCreate: true,
			check: func(t *testing.T, input map[string]any) {
				recovery, _ := input["recovery"].(map[string]any)
				target, err := time.Parse(time.RFC3339, recovery["targetTime"].(string))
				if err != nil {
					t.Fatal(err)
				}
				if age := time.Since(target); age < 9*time.Minute+50*time.Second || age > 10*time.Minute+30*time.Second {
					t.Errorf("default target is %v old, want about 10 minutes", age)
				}
			},
		},
		{name: "default earlier than the backup window explains itself", source: strings.Replace(source, `"2026-11-10T00:00:00Z"`, `"2099-01-01T00:00:00Z"`, 1), wantError: "default clone point"},
		{name: "invalid time is rejected before the source is read", extra: []string{"--at", "2026-11-12T10:00:00.5Z"}, wantError: "whole seconds"},
		{name: "source cannot be its own target", target: "orders", extra: []string{"--at", "2026-11-12T10:00:00Z"}, wantError: "different name"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stdin, err := os.Open(os.DevNull)
			if err != nil {
				t.Fatal(err)
			}
			previous := os.Stdin
			os.Stdin = stdin
			t.Cleanup(func() { os.Stdin = previous; _ = stdin.Close() })
			var operations []string
			var created map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string         `json:"operationName"`
					Variables     map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				w.Header().Set("Content-Type", "application/json")
				switch request.OperationName {
				case "GetTeamPostgreses":
					_, _ = w.Write([]byte(`{"data":{"team":{"postgreses":{"nodes":[{"name":"orders","teamEnvironment":{"environment":{"name":"dev-gcp"}}}],"pageInfo":{"hasNextPage":false}}}}}`))
				case "GetCloneSource":
					_, _ = w.Write([]byte(`{"data":{"team":{"environment":{"postgres":` + tt.source + `}}}}`))
				case "CreatePostgres":
					created, _ = request.Variables["input"].(map[string]any)
					_, _ = w.Write([]byte(`{"data":{"createPostgres":{"postgres":{"name":"orders-restored"}}}}`))
				default:
					t.Errorf("unexpected operation %s", request.OperationName)
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
			app, global, err := naistrix.NewApplication("postgres-restore-test", "Test Postgres restore", "test", naistrix.ApplicationWithWriter(&output))
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
			target := tt.target
			if target == "" {
				target = "orders-restored"
			}
			args := append([]string{"--config", config, "--no-colors", "postgres", "clone", "orders", target, "-t", "my-team", "-e", "dev-gcp", "--yes"}, tt.extra...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if (created != nil) != tt.wantCreate {
				t.Fatalf("created = %v, want create %t (operations %v)", created, tt.wantCreate, operations)
			}
			if tt.wantError != "" && tt.source == "" || tt.target == "orders" {
				for _, operation := range operations {
					if operation == "GetCloneSource" || operation == "CreatePostgres" {
						t.Errorf("rejected input must not reach %s (operations %v)", operation, operations)
					}
				}
			}
			if tt.check != nil {
				tt.check(t, created)
			}
			if tt.wantCreate && !strings.Contains(output.String(), "Started cloning postgres") {
				t.Errorf("output = %q", output.String())
			}
		})
	}
}

func TestGetShowsCloneProgress(t *testing.T) {
	for _, tt := range []struct {
		name, recovery string
		want           []string
	}{
		{"in progress", `{"sourcePostgres":"orders","targetTime":"2026-11-12T10:00:00Z","completed":false,"failureMessage":null}`, []string{"Cloned from:", "orders", "Clone point:", "In progress"}},
		{"completed", `{"sourcePostgres":"orders","targetTime":"2026-11-12T10:00:00Z","completed":true,"failureMessage":null}`, []string{"Completed"}},
		{"failed", `{"sourcePostgres":"orders","targetTime":"2026-11-12T10:00:00Z","completed":false,"failureMessage":"archive missing"}`, []string{"Failed: archive missing"}},
		{"not restored", `null`, []string{"Version:"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string `json:"operationName"`
				}
				_ = json.NewDecoder(r.Body).Decode(&request)
				w.Header().Set("Content-Type", "application/json")
				if request.OperationName == "GetTeamPostgreses" {
					_, _ = w.Write([]byte(`{"data":{"team":{"postgreses":{"nodes":[{"name":"restored","teamEnvironment":{"environment":{"name":"dev-gcp"}}}],"pageInfo":{"hasNextPage":false}}}}}`))
					return
				}
				_, _ = w.Write([]byte(`{"data":{"team":{"environment":{"postgres":{"name":"restored","majorVersion":"18","state":"AVAILABLE","highAvailability":false,"resources":{},"labels":[],"recovery":` + tt.recovery + `}}}}}}`))
			}))
			defer server.Close()
			t.Setenv("NAIS_API_LOCAL_HOST", strings.TrimPrefix(server.URL, "http://"))
			t.Setenv("HOME", t.TempDir())
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, format := range []string{"", "json"} {
				var output bytes.Buffer
				app, global, err := naistrix.NewApplication("postgres-get-test", "Test", "test", naistrix.ApplicationWithWriter(&output))
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
				args := []string{"--config", config, "--no-colors", "postgres", "get", "restored", "-t", "my-team", "-e", "dev-gcp"}
				if format != "" {
					args = append(args, "--output", format)
				}
				if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
					t.Fatal(err)
				}
				if format == "json" {
					var detail postgresDetails
					if err := json.Unmarshal(output.Bytes(), &detail); err != nil {
						t.Fatalf("invalid JSON %q: %v", output.String(), err)
					}
					if (detail.Recovery != nil) != (tt.recovery != "null") {
						t.Errorf("recovery = %+v for %s", detail.Recovery, tt.recovery)
					}
					continue
				}
				for _, want := range tt.want {
					if !strings.Contains(output.String(), want) {
						t.Errorf("output missing %q:\n%s", want, output.String())
					}
				}
				if tt.recovery == "null" && strings.Contains(output.String(), "Clone") {
					t.Errorf("a database that was not restored must not show restore details:\n%s", output.String())
				}
			}
		})
	}
}

// Lord Howe Island shifts its clocks by 30 minutes, so a repeated or skipped local time cannot be
// found by trying one-hour offsets.
func TestParseLocalTimeWithHalfHourDaylightSaving(t *testing.T) {
	lordHowe, err := time.LoadLocation("Australia/Lord_Howe")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name, value, wantError string
		want                   time.Time
	}{
		{name: "repeated half hour resolves to the earlier occurrence", value: "2026-04-05 01:45", want: time.Date(2026, 4, 4, 14, 45, 0, 0, time.UTC)},
		{name: "end of the repeated half hour is still the earlier occurrence", value: "2026-04-05 01:59:59", want: time.Date(2026, 4, 4, 14, 59, 59, 0, time.UTC)},
		{name: "first time after the repeat is unambiguous", value: "2026-04-05 02:00", want: time.Date(2026, 4, 4, 15, 30, 0, 0, time.UTC)},
		{name: "before the repeat is unambiguous", value: "2026-04-05 01:15", want: time.Date(2026, 4, 4, 14, 15, 0, 0, time.UTC)},
		{name: "half hour skipped in spring is rejected", value: "2026-10-04 02:15", wantError: "does not exist"},
		{name: "first time after the skipped half hour", value: "2026-10-04 02:30", want: time.Date(2026, 10, 3, 15, 30, 0, 0, time.UTC)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLocalTime(tt.value, lordHowe)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil || !got.Equal(tt.want) {
				t.Fatalf("got %v, %v; want %v", got.UTC(), err, tt.want)
			}
			if _, _, err := resolveCloneTime(tt.value, now, lordHowe); err != nil {
				t.Errorf("resolveCloneTime: %v", err)
			}
		})
	}
}
