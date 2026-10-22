package application

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/naistrix"
)

func TestPostgresCommandMigration(t *testing.T) {
	for _, tt := range []struct {
		name     string
		args     []string
		want     []string
		rejected bool
	}{
		{
			name: "postgres manages Nais Postgres",
			args: []string{"postgres", "--help"},
			want: []string{"Manage Nais Postgres instances.", "branch", "create"},
		},
		{
			name: "postgres completes Nais Postgres commands",
			args: []string{"__complete", "postgres", ""},
			want: []string{"branch\t", "create\t", "get\t"},
		},
		{
			name: "Cloud SQL remains available",
			args: []string{"cloudsql", "--help"},
			want: []string{"Manage Google Cloud SQL instances.", "password", "users"},
		},
		{
			name:     "alpha postgres is removed",
			args:     []string{"alpha", "postgres", "list"},
			rejected: true,
		},
		{
			name:     "pg Cloud SQL alias is removed",
			args:     []string{"pg", "list"},
			rejected: true,
		},
		{
			name:     "Cloud SQL commands are not exposed through postgres",
			args:     []string{"postgres", "password", "rotate"},
			rejected: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			app, _, err := New(&output)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--config", config, "--no-colors"}, tt.args...)
			err = app.Run(naistrix.RunWithArgs(args))
			if tt.rejected {
				if err == nil {
					t.Fatal("removed command was accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.want {
				if !strings.Contains(output.String(), want) {
					t.Errorf("output missing %q: %s", want, output.String())
				}
			}
			if strings.Contains(strings.ToLower(output.String()), "experimental") {
				t.Errorf("stable command still marked experimental: %s", output.String())
			}
		})
	}
}
