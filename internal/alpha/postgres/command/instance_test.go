package command

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func TestInstanceCommands(t *testing.T) {
	pg := Postgres(&flags.GlobalFlags{})
	for _, name := range []string{"create", "update", "delete"} {
		var found *naistrix.Command
		for _, command := range pg.SubCommands {
			if command.Name == name {
				found = command
				break
			}
		}
		if found == nil {
			t.Errorf("%s command not registered", name)
			continue
		}
		if len(found.Args) != 1 || found.Args[0].Name != "postgres" {
			t.Errorf("%s args = %v, want one postgres argument", name, found.Args)
		}
		if (found.AutoCompleteFunc != nil) != (name != "create") {
			t.Errorf("%s completion registration does not match name contract", name)
		}
	}
}

func TestInstanceFlagSuggestions(t *testing.T) {
	for _, tc := range []struct {
		name string
		flag naistrix.FlagAutoCompleter
		want []string
	}{
		{"version", new(flag.PostgresVersion), []string{"18"}},
		{"update HA", new(flag.HighAvailability), []string{"true", "false"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.flag.AutoComplete(context.Background(), nil, "", nil)
			if !slices.Equal(got, tc.want) {
				t.Errorf("suggestions = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUpdateFields(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags flag.InstanceUpdate
		want  string
		err   bool
	}{
		{name: "no changes", err: true},
		{name: "invalid HA", flags: flag.InstanceUpdate{HighAvailability: "yes"}, err: true},
		{name: "enable HA", flags: flag.InstanceUpdate{HighAvailability: "true"}, want: `"highAvailability":true`},
		{name: "disable HA", flags: flag.InstanceUpdate{HighAvailability: "false"}, want: `"highAvailability":false`},
		{name: "CPU only", flags: flag.InstanceUpdate{CPU: "100m"}, want: `"cpu":"100m"`},
		{name: "memory only", flags: flag.InstanceUpdate{Memory: "512Mi"}, want: `"memory":"512Mi"`},
		{name: "disk only", flags: flag.InstanceUpdate{DiskSize: "10Gi"}, want: `"diskSize":"10Gi"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input, err := updateFields(&tc.flags)
			if (err != nil) != tc.err {
				t.Fatalf("error = %v, want error %v", err, tc.err)
			}
			if err != nil {
				return
			}
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), tc.want) {
				t.Errorf("request %s does not include %s", data, tc.want)
			}
			for _, field := range []string{`"highAvailability"`, `"cpu"`, `"memory"`, `"diskSize"`} {
				if field != strings.Split(tc.want, ":")[0] && strings.Contains(string(data), field) {
					t.Errorf("request %s unexpectedly includes %s", data, field)
				}
			}
		})
	}
}

func TestDeleteRequiresConfirmationBeforeRequest(t *testing.T) {
	target := gql.DeletePostgresInput{Name: "db", TeamSlug: "team", EnvironmentName: "dev"}
	for _, tc := range []struct {
		name, wantError string
		yes, accept     bool
		confirmError    error
		wantCalls       int
	}{
		{name: "cancel", wantError: "cancelled by user"},
		{name: "prompt failure", confirmError: errors.New("no terminal"), wantError: "no terminal"},
		{name: "accept", accept: true, wantCalls: 1},
		{name: "skip prompt with yes", yes: true, wantCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			requested, err := confirmAndDeleteInstance(context.Background(), target, tc.yes,
				func(prompt string) (bool, error) {
					if tc.yes {
						t.Fatal("prompt called with --yes")
					}
					for _, detail := range []string{target.Name, target.TeamSlug, target.EnvironmentName, "Irreversibly", "all its branches/data"} {
						if !strings.Contains(prompt, detail) {
							t.Errorf("prompt %q missing %q", prompt, detail)
						}
					}
					return tc.accept, tc.confirmError
				},
				func(_ context.Context, got gql.DeletePostgresInput) (bool, error) {
					calls++
					if got != target {
						t.Errorf("request = %+v, want %+v", got, target)
					}
					return true, nil
				})
			if calls != tc.wantCalls || requested != (tc.wantCalls == 1) {
				t.Errorf("calls = %d, accepted = %v; want %d calls", calls, requested, tc.wantCalls)
			}
			if tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Errorf("error = %v, want %q", err, tc.wantError)
			} else if tc.wantError == "" && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreateOptionalFields(t *testing.T) {
	if enabledHA(false) != nil || optional("") != nil {
		t.Fatal("omitted create settings must use API defaults")
	}
	if value := enabledHA(true); value == nil || !*value {
		t.Fatal("enabled HA must be sent")
	}
}
