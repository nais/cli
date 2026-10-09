package command

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/member/command/flag"
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
}

func TestSetRoleAcceptsCaseInsensitiveRoles(t *testing.T) {
	for _, role := range []string{"member", "MEMBER", "Member", "owner", "OWNER", "Owner"} {
		t.Run(role, func(t *testing.T) {
			cmd := setRole(&flag.Member{})
			cmd.Flags = nil
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
	cmd := setRole(&flag.Member{})
	got, _ := cmd.AutoCompleteFunc(context.Background(), &naistrix.Arguments{}, "")
	if !slices.Equal(got, cmd.Args[0].Choices) {
		t.Errorf("role completions = %v, want %v", got, cmd.Args[0].Choices)
	}
}
