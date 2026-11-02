package application

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAuthStatusJSON(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script to stub gcloud")
	}
	for _, tt := range []struct {
		name, script, status string
		wantError            bool
	}{
		{name: "gcloud not installed", status: "not_installed"},
		{name: "gcloud check failure", script: "#!/bin/sh\nexit 1\n", status: "check_failed", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", home)
			t.Setenv("PATH", home)
			t.Setenv("DO_NOT_TRACK", "1")
			t.Setenv("NAIS_API_LOCAL_HOST", "localhost:4242")
			t.Setenv("NAIS_API_LOCAL_EMAIL", "nais@example.com")
			if tt.script != "" {
				if err := os.WriteFile(filepath.Join(home, "gcloud"), []byte(tt.script), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			previousArgs := os.Args
			t.Cleanup(func() { os.Args = previousArgs })
			os.Args = []string{"nais", "--config", filepath.Join(home, "config.yaml"), "--no-colors", "auth", "status", "-o", "json", "--confirm-login=false"}
			var stdout bytes.Buffer
			err := Run(t.Context(), &stdout)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}
			var report struct {
				Nais, Gcloud struct{ Status string }
			}
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("stdout is not a single JSON document: %v\n%s", err, stdout.String())
			}
			if report.Nais.Status != "authenticated" || report.Gcloud.Status != tt.status {
				t.Errorf("unexpected report: %s", stdout.String())
			}
			if tt.wantError && !strings.Contains(err.Error(), "Google Cloud authentication status:") {
				t.Errorf("missing provider details in returned error: %v", err)
			}
		})
	}
}
