package command

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/nais/naistrix"
)

func TestConfigValueArgumentChoices(t *testing.T) {
	cmd := set()
	choices := cmd.Args[1].Choices
	if want := []string{"true", "false"}; !slices.Equal(choices, want) {
		t.Errorf("boolean choices = %v, want %v", choices, want)
	}
}

func TestConfigValueCaseInsensitiveChoices(t *testing.T) {
	for _, value := range []string{"true", "TRUE", "True", "false", "FALSE", "False", "1", "0", "t", "f", "invalid"} {
		t.Run(value, func(t *testing.T) {
			cmd := set()
			var got string
			cmd.RunFunc = func(_ context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
				got = args.Get("value")
				return nil
			}
			app, _, err := naistrix.NewApplication("test", "Test", "v0.0.0")
			if err != nil {
				t.Fatal(err)
			}
			if err := app.AddCommand(cmd); err != nil {
				t.Fatal(err)
			}
			err = app.Run(naistrix.RunWithArgs([]string{"set", "AutoConnect", value}))
			want := strings.ToLower(value)
			if want != "true" && want != "false" {
				if err == nil || !strings.Contains(err.Error(), "invalid value") {
					t.Fatalf("expected invalid value error, got %v", err)
				}
				if got != "" {
					t.Error("command ran with an invalid boolean")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("value = %q, want %q", got, want)
			}
		})
	}
}

func TestConfigSettingArgumentsRemainUnrestricted(t *testing.T) {
	for _, cmd := range []*naistrix.Command{get(), set()} {
		if len(cmd.Args[0].Choices) != 0 {
			t.Errorf("%s setting choices must remain empty to allow case-insensitive and hidden settings", cmd.Name)
		}
	}
}
