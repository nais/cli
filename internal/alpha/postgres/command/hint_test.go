package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostgresGetCommandLine(t *testing.T) {
	for _, tt := range []struct {
		name, config, envTeam, envEnvironment, team, environment, postgres, wantScope string
		custom, emptyPath, missing, unreadable                                        bool
	}{
		{name: "no defaults", emptyPath: true, wantScope: " -t my-team -e dev-gcp"},
		{name: "matching file defaults", config: "team: my-team\nenvironment: dev-gcp\n"},
		{name: "only matching team default", config: "team: my-team\n", wantScope: " -e dev-gcp"},
		{name: "only matching environment default", config: "environment: dev-gcp\n", wantScope: " -t my-team"},
		{name: "explicit overrides differ from file defaults", config: "team: other-team\nenvironment: prod-gcp\n", wantScope: " -t my-team -e dev-gcp"},
		{name: "environment defaults override stored defaults", config: "team: other-team\nenvironment: prod-gcp\n", envTeam: "my-team", envEnvironment: "dev-gcp"},
		{name: "selected environment without default remains explicit", emptyPath: true, envTeam: "my-team", wantScope: " -e dev-gcp"},
		{name: "custom file defaults require config path", config: "team: my-team\nenvironment: dev-gcp\n", custom: true},
		{name: "missing config still permits environment defaults", custom: true, missing: true, envTeam: "my-team", envEnvironment: "dev-gcp"},
		{name: "missing config has no file defaults", custom: true, missing: true, wantScope: " -t my-team -e dev-gcp"},
		{name: "unreadable config falls back to explicit scope", custom: true, unreadable: true, envTeam: "my-team", envEnvironment: "dev-gcp", wantScope: " -t my-team -e dev-gcp"},
		{name: "malformed config falls back to explicit scope", config: "team: [invalid", custom: true, envTeam: "my-team", envEnvironment: "dev-gcp", wantScope: " -t my-team -e dev-gcp"},
		{name: "shell metacharacters are quoted", emptyPath: true, postgres: "db'$(touch file)", team: "team name", environment: "env$HOME", wantScope: " -t 'team name' -e 'env$HOME'"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("APPDATA", t.TempDir())
			t.Setenv("NAIS_TEAM", tt.envTeam)
			t.Setenv("NAIS_ENVIRONMENT", tt.envEnvironment)
			configDir, err := os.UserConfigDir()
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(configDir, "nais", "config.yaml")
			if tt.custom {
				file = filepath.Join(t.TempDir(), "custom '$(touch file).yaml")
			}
			if tt.emptyPath {
				file = ""
			} else if !tt.missing {
				if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
					t.Fatal(err)
				}
				if tt.unreadable {
					if err := os.Mkdir(file, 0o700); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(file, []byte(tt.config), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			name, team, environment := tt.postgres, tt.team, tt.environment
			if name == "" {
				name = "orders"
			}
			if team == "" {
				team = "my-team"
			}
			if environment == "" {
				environment = "dev-gcp"
			}
			want := "nais alpha postgres get orders"
			if tt.postgres != "" {
				want = "nais alpha postgres get 'db'\"'\"'$(touch file)'"
			}
			if tt.custom {
				want += " --config '" + filepath.Join(filepath.Dir(file), "custom '\"'\"'$(touch file).yaml") + "'"
			}
			want += tt.wantScope
			got := postgresGetCommandLine(name, team, environment, file)
			if got != want {
				t.Errorf("hint = %q; want %q", got, want)
			}
			if strings.Contains(got, "postgres status") {
				t.Error("hint refers to removed top-level status command")
			}
		})
	}
}
