package application

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/naistrix"
)

func TestDefaultsDocumentation(t *testing.T) {
	var output bytes.Buffer
	app, flags, err := New(&output)
	if err != nil {
		t.Fatal(err)
	}
	if flags.Team != "" || flags.Environment != "" {
		t.Fatal("documenting defaults must not select a team or environment")
	}
	if !flags.ConfirmLogin {
		t.Fatal("documenting defaults must not disable login confirmations")
	}

	dir := t.TempDir()
	if err := app.GenerateDocs(naistrix.GenerateDocsWithTargetDir(dir)); err != nil {
		t.Fatal(err)
	}
	docs, err := os.ReadFile(filepath.Join(dir, "configuration.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"\n### team\n", "\n### environment\n", "\n### confirm-login\n", "Nais Console",
		"Default Nais team", "Default Nais environment", "nais_defaults_describe",
		"nais defaults set team <team>", "nais defaults set environment <env>",
		"nais app list --team <team>", "nais app list --environment <env>",
		"nais defaults unset team", "nais defaults unset environment",
		"Ask before starting required Nais or Google Cloud login flows.",
		"Defaults to", "nais defaults set confirm-login false",
		"--confirm-login=false", "--confirm-login=true",
		"nais defaults unset confirm-login",
	} {
		if !strings.Contains(string(docs), text) {
			t.Errorf("generated configuration documentation is missing %q", text)
		}
	}

	output.Reset()
	if err := app.Run(naistrix.RunWithArgs([]string{"--config", filepath.Join(dir, "config.yaml"), "defaults", "list"})); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"team", "environment", "confirm-login", "Default Nais team", "Default Nais environment", "Ask before starting required Nais or Google Cloud login flows.", "(not set)"} {
		if !strings.Contains(output.String(), text) {
			t.Errorf("defaults list is missing %q", text)
		}
	}
	for _, text := range []string{"Nais Console", "nais defaults set team <team>", "nais defaults set environment <env>", "nais defaults set confirm-login false"} {
		if strings.Contains(output.String(), text) {
			t.Errorf("defaults list should show short titles, not detailed documentation %q", text)
		}
	}
}

func TestDefaultsDescribeDocumentation(t *testing.T) {
	for _, tt := range []struct {
		key  string
		want []string
	}{
		{
			key: "confirm-login",
			want: []string{
				"Ask before starting required Nais or Google Cloud login flows.", "Defaults to true.",
				"nais defaults set confirm-login false", "nais defaults set confirm-login true",
				"--confirm-login=false", "--confirm-login=true",
				"nais defaults unset confirm-login", "Configured value: (not set)",
			},
		},
		{
			key: "team",
			want: []string{
				"Default Nais team", "Nais Console",
				"nais defaults set team <team>", "nais app list --team <team>",
				"nais defaults unset team", "Configured value: (not set)",
			},
		},
		{
			key: "environment",
			want: []string{
				"Default Nais environment", "Nais environment",
				"nais defaults set environment <env>", "nais app list --environment <env>",
				"nais defaults unset environment", "Configured value: (not set)",
			},
		},
	} {
		t.Run(tt.key, func(t *testing.T) {
			var output bytes.Buffer
			app, _, err := New(&output)
			if err != nil {
				t.Fatal(err)
			}
			if err := app.Run(naistrix.RunWithArgs([]string{
				"--config", filepath.Join(t.TempDir(), "config.yaml"), "--no-colors", "defaults", "describe", tt.key,
			})); err != nil {
				t.Fatal(err)
			}
			for _, text := range tt.want {
				if !strings.Contains(output.String(), text) {
					t.Errorf("defaults describe %s is missing %q", tt.key, text)
				}
			}
		})
	}
}
