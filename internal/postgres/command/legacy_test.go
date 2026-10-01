package command

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func TestLegacyCommandWarnsAndRunsOriginal(t *testing.T) {
	for _, tt := range []struct {
		name        string
		path        string
		replacement string
		want        string
	}{
		{"replacement", "postgres prepare", "nais cloudsql prepare", "Warning: nais postgres prepare is deprecated and will be removed in a future release; use nais cloudsql prepare instead; the legacy in-cluster access path will be removed.\n"},
		{"no replacement", "postgres grant", "", "Warning: nais postgres grant is deprecated and will be removed in a future release; the legacy in-cluster access path will be removed.\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			originalErr := errors.New("original error")
			ctx := context.Background()
			args := &naistrix.Arguments{}
			var stdout bytes.Buffer
			out := naistrix.NewOutputWriter(&stdout, nil)
			called := false
			cmd := legacyCommand(tt.path, tt.replacement, &naistrix.Command{
				Name: "prepare",
				RunFunc: func(gotCtx context.Context, gotArgs *naistrix.Arguments, gotOut *naistrix.OutputWriter) error {
					called = true
					if gotCtx != ctx || gotArgs != args || gotOut != out {
						t.Error("original command received different inputs")
					}
					return originalErr
				},
			})

			previous := os.Stderr
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stderr = w
			gotErr := cmd.RunFunc(ctx, args, out)
			os.Stderr = previous
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			warning, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.Close(); err != nil {
				t.Fatal(err)
			}
			if string(warning) != tt.want {
				t.Errorf("stderr = %q, want %q", warning, tt.want)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if !called || !errors.Is(gotErr, originalErr) {
				t.Errorf("original called = %v, error = %v; want original error", called, gotErr)
			}
		})
	}
}

func TestLegacyPostgresLeavesRemainRunnable(t *testing.T) {
	root := Postgres(&flags.GlobalFlags{})
	var paths []string
	var walk func(*naistrix.Command, string)
	walk = func(cmd *naistrix.Command, prefix string) {
		path := strings.TrimSpace(prefix + " " + cmd.Name)
		if len(cmd.SubCommands) == 0 {
			paths = append(paths, path)
			if cmd.Deprecated != nil || cmd.RunFunc == nil || !strings.Contains(cmd.Title, "(DEPRECATED)") {
				t.Errorf("%s must execute its RunFunc without naistrix deprecation interception", path)
			}
		}
		for _, child := range cmd.SubCommands {
			walk(child, path)
		}
	}
	walk(root, "")
	want := []string{
		"postgres list", "postgres migrate setup", "postgres migrate promote", "postgres migrate finalize", "postgres migrate rollback",
		"postgres password rotate", "postgres users add", "postgres users drop", "postgres users list",
		"postgres enable-audit", "postgres verify-audit", "postgres grant", "postgres prepare", "postgres proxy", "postgres psql", "postgres revoke",
	}
	if strings.Join(paths, ",") != strings.Join(want, ",") {
		t.Errorf("leaves = %v, want %v", paths, want)
	}
}
