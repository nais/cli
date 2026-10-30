package gcloud

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateAndGetUserLogin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script to stub gcloud")
	}
	for _, tt := range []struct {
		name, account, credentials string
		adc, fail, enforceNais     bool
		wantLogin, wantError       bool
	}{
		{name: "missing account", wantLogin: true, wantError: true},
		{name: "missing ADC", account: "user@example.com", wantLogin: true, wantError: true},
		{name: "valid", account: "user@example.com", adc: true},
		{name: "explicit credentials", account: "user@example.com", credentials: "/external/credentials.json"},
		{name: "wrong domain", account: "user@example.com", enforceNais: true, wantError: true},
		{name: "gcloud failure", fail: true, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("PATH", home)
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.credentials)
			if tt.credentials == "" {
				if err := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS"); err != nil {
					t.Fatal(err)
				}
			}
			script := "#!/bin/sh\nprintf '%s\\n' '" + tt.account + "'\n"
			if tt.fail {
				script = "#!/bin/sh\nexit 1\n"
			}
			if err := os.WriteFile(filepath.Join(home, "gcloud"), []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			if tt.adc {
				dir := filepath.Join(home, ".config", "gcloud")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "application_default_credentials.json"), []byte("{}"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			user, err := ValidateAndGetUserLogin(t.Context(), tt.enforceNais)
			if (err != nil) != tt.wantError || errors.Is(err, ErrNeedsLogin) != tt.wantLogin {
				t.Fatalf("error = %v, wantError = %v, wantLogin = %v", err, tt.wantError, tt.wantLogin)
			}
			if err == nil && user != tt.account {
				t.Errorf("user = %q, want %q", user, tt.account)
			}
		})
	}
}
