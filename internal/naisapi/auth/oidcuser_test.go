package auth

import (
	"path/filepath"
	"testing"

	gokeyring "github.com/zalando/go-keyring"
)

func setTestConfigPath(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	oldPath := ConfigFilePath
	ConfigFilePath = &path
	t.Cleanup(func() { ConfigFilePath = oldPath })
}

// setupCredentialStore gives the test an empty mock keyring and its own config directory.
func setupCredentialStore(t *testing.T) {
	t.Helper()
	gokeyring.MockInit()
	t.Cleanup(gokeyring.MockInit)
	setTestConfigPath(t, t.TempDir())
}
