package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/member/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func TestSetRoleArgumentChoices(t *testing.T) {
	cmd := setRole(&flag.Member{})
	if got, want := cmd.Args[0].Choices, []string{"member", "owner"}; !slices.Equal(got, want) {
		t.Errorf("role choices = %v, want %v", got, want)
	}
	if len(cmd.Args[1].Choices) != 0 {
		t.Error("member emails must not have static choices")
	}
	if cmd.Args[1].ChoicesFunc == nil {
		t.Error("member emails must have dynamic choices")
	}
	if cmd.AutoCompleteFunc != nil {
		t.Error("completion must use argument choices rather than a custom function")
	}
}

func TestSetRoleMemberChoicesRequireTeam(t *testing.T) {
	cmd := setRole(&flag.Member{GlobalFlags: &flags.GlobalFlags{AdditionalFlags: &flags.AdditionalFlags{}}})
	_, err := cmd.Args[1].ChoicesFunc(context.Background(), nil)
	if !errors.Is(err, validation.ErrMissingTeam) {
		t.Errorf("member choices error = %v, want ErrMissingTeam", err)
	}
}

func TestSetRoleAcceptsCaseInsensitiveRoles(t *testing.T) {
	for _, role := range []string{"member", "MEMBER", "Member", "owner", "OWNER", "Owner"} {
		t.Run(role, func(t *testing.T) {
			cmd := setRole(&flag.Member{})
			cmd.Flags = nil
			cmd.Args[1].ChoicesFunc = func(context.Context, *naistrix.Arguments) ([]string, error) {
				return []string{"user@example.com"}, nil
			}
			var got string
			cmd.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
				got = args.Get("role")
				return nil
			}
			app, _, err := naistrix.NewApplication("test", "Test", "v0.0.0")
			if err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(cmd); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs([]string{"set-role", role, "user@example.com"})); err != nil {
				t.Fatal(err)
			}
			if want := strings.ToLower(role); got != want {
				t.Errorf("role = %q, want %q", got, want)
			}
		})
	}
}

func TestSetRoleCompletionMatchesChoices(t *testing.T) {
	for _, tt := range []struct {
		name        string
		args        []string
		want        string
		memberCalls int
	}{
		{name: "roles", args: []string{""}, want: "member\nowner\n_activeHelp_ Role to assign to the team member\n:4\n"},
		{name: "uppercase role prefix", args: []string{"O"}, want: "Owner\n_activeHelp_ Role to assign to the team member\n:4\n"},
		{name: "members", args: []string{"OWNER", ""}, want: "user@example.com\nother@example.com\n_activeHelp_ Member of \"my-team\" to assign the \"owner\" role to\n:4\n", memberCalls: 1},
		{name: "member prefix", args: []string{"owner", "u"}, want: "user@example.com\n_activeHelp_ Member of \"my-team\" to assign the \"owner\" role to\n:4\n", memberCalls: 1},
		{name: "all arguments supplied", args: []string{"owner", "user@example.com", ""}, want: ":4\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := setRole(&flag.Member{GlobalFlags: &flags.GlobalFlags{AdditionalFlags: &flags.AdditionalFlags{Team: "my-team"}}})
			cmd.Flags = nil
			calls := 0
			cmd.Args[1].ChoicesFunc = func(_ context.Context, args *naistrix.Arguments) ([]string, error) {
				calls++
				if got := args.Get("role"); got != "owner" {
					t.Errorf("role passed to member choices = %q, want owner", got)
				}
				return []string{"user@example.com", "other@example.com"}, nil
			}
			cmd.RunFunc = func(context.Context, *naistrix.Arguments, *naistrix.OutputWriter) error {
				t.Error("completion must not execute the command")
				return nil
			}
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test", "v0.0.0", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			if err := app.AddCommand(cmd); err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs(append([]string{"__complete", "set-role"}, tt.args...))); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != tt.want {
				t.Errorf("completion = %q, want %q", got, tt.want)
			}
			if calls != tt.memberCalls {
				t.Errorf("member choices calls = %d, want %d", calls, tt.memberCalls)
			}
		})
	}
}

func TestSetRoleDynamicMemberChoices(t *testing.T) {
	for _, tt := range []struct {
		name, email, response, wantError string
		completion, omitTeam             bool
		wantOperations                   []string
	}{
		{name: "existing member", email: "user@example.com", wantOperations: []string{"TeamMembers", "SetRole"}},
		{name: "unknown member", email: "unknown@example.com", wantError: "invalid value", wantOperations: []string{"TeamMembers"}},
		{name: "empty team", email: "user@example.com", response: `{"data":{"team":{"members":{"nodes":[]}}}}`, wantError: "no choices available", wantOperations: []string{"TeamMembers"}},
		{name: "API failure", email: "user@example.com", response: `{"errors":[{"message":"access denied"}]}`, wantError: "unable to fetch existing team members: input: access denied", wantOperations: []string{"TeamMembers"}},
		{name: "missing team", email: "user@example.com", omitTeam: true, wantError: "missing required team"},
		{name: "member completion", completion: true, wantOperations: []string{"TeamMembers"}},
		{name: "completion API failure is visible", completion: true, response: `{"errors":[{"message":"access denied"}]}`, wantError: "unable to fetch existing team members: input: access denied", wantOperations: []string{"TeamMembers"}},
		{name: "completion empty team is visible", completion: true, response: `{"data":{"team":{"members":{"nodes":[]}}}}`, wantError: "no choices available", wantOperations: []string{"TeamMembers"}},
		{name: "completion missing team is visible", completion: true, omitTeam: true, wantError: "missing required team"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var operations []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					OperationName string                             `json:"operationName"`
					Variables     struct{ Slug, Email, Role string } `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decoding request: %v", err)
				}
				operations = append(operations, request.OperationName)
				if request.Variables.Slug != "my-team" {
					t.Errorf("team slug = %q, want my-team", request.Variables.Slug)
				}
				response := tt.response
				switch request.OperationName {
				case "TeamMembers":
					if response == "" {
						response = `{"data":{"team":{"members":{"nodes":[{"role":"MEMBER","user":{"name":"User","email":"user@example.com"}}]}}}}`
					}
				case "SetRole":
					if request.Variables.Email != tt.email || request.Variables.Role != "OWNER" {
						t.Errorf("unexpected role mutation variables: %+v", request.Variables)
					}
					response = `{"data":{"setTeamMemberRole":{"member":{"role":"OWNER","user":{"email":"user@example.com"}}}}}`
				default:
					t.Errorf("unexpected operation: %s", request.OperationName)
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
			app, global, err := naistrix.NewApplication("test", "Test", "v0.0.0", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(t.TempDir(), "config.yaml")
			additional := &flags.AdditionalFlags{}
			if err := app.AddGlobalFlags(additional); err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(Members(&flags.GlobalFlags{GlobalFlags: global, AdditionalFlags: additional})); err != nil {
				t.Fatal(err)
			}
			args := []string{"members", "set-role", "OWNER"}
			if !tt.omitTeam {
				args = append(args, "--team", "my-team")
			}
			if tt.completion {
				args = append([]string{"__complete"}, args...)
				args = append(args, "")
			} else {
				args = append(args, tt.email)
			}
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.completion && tt.wantError != "" {
				if err != nil {
					t.Fatal(err)
				}
				got := output.String()
				if !strings.HasPrefix(got, "_activeHelp_ Argument completion failed:") || !strings.Contains(got, tt.wantError) || !strings.HasSuffix(got, "\n:4\n") || strings.Count(got, "\n") != 2 {
					t.Errorf("completion = %q, want a single active help error containing %q and NoFileComp directive", got, tt.wantError)
				}
			} else if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if tt.completion && tt.wantError == "" && output.String() != "user@example.com\n_activeHelp_ Member of \"my-team\" to assign the \"owner\" role to\n:4\n" {
				t.Errorf("completion = %q, want member email, active help, and NoFileComp directive", output.String())
			}
			if !slices.Equal(operations, tt.wantOperations) {
				t.Errorf("API operations = %v, want %v", operations, tt.wantOperations)
			}
		})
	}
}
