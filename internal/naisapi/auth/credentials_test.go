package auth

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/nais/cli/internal/keyring"
	gokeyring "github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

// setTestUserConfigHome isolates the credentials lock, which lives in the user config directory.
func setTestUserConfigHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
}

func setTestConfigPath(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "config.yaml")
	oldPath := ConfigFilePath
	ConfigFilePath = &path
	t.Cleanup(func() { ConfigFilePath = oldPath })
}

func setupCredentialStore(t *testing.T) {
	t.Helper()
	gokeyring.MockInit()
	t.Cleanup(gokeyring.MockInit)
	setTestUserConfigHome(t)
	setTestConfigPath(t, t.TempDir())
}

// storeTestUser stores a session with an opaque ID token, for tests that never validate it.
func storeTestUser(t *testing.T) {
	t.Helper()
	tok := (&oauth2.Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}).WithExtra(map[string]any{"id_token": "id-token"})
	if _, err := storeOIDCUser(t.Context(), tok, "console.example.com"); err != nil {
		t.Fatal(err)
	}
}

func testEncryptionKey(t *testing.T) string {
	t.Helper()
	key, err := keyring.Get()
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func readTestCredentialsFile(t *testing.T) []byte {
	t.Helper()
	path, err := getCredentialsFilePath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// skipWithoutUnixPermissions skips chmod-based tests: Windows ignores chmod, and root bypasses it.
func skipWithoutUnixPermissions(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("chmod does not enforce permissions on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}
}

func TestConfigDirectoriesShareEncryptionKey(t *testing.T) {
	setupCredentialStore(t)
	dirA, dirB := t.TempDir(), t.TempDir()

	setTestConfigPath(t, dirA)
	storeTestUser(t)
	key := testEncryptionKey(t)
	storeTestUser(t)
	setTestConfigPath(t, dirB)
	storeTestUser(t)

	if testEncryptionKey(t) != key {
		t.Fatal("storing credentials rotated the shared encryption key")
	}

	// Logging out of A also logs out B, since it deletes the shared key.
	setTestConfigPath(t, dirA)
	for range 2 {
		if err := deleteOIDCUser(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{dirA, dirB} {
		setTestConfigPath(t, dir)
		if _, err := getOIDCUser(t.Context()); !errors.Is(err, ErrNeedsOIDCLogin) {
			t.Fatalf("getOIDCUser() after logout = %v, want ErrNeedsOIDCLogin", err)
		}
	}

	// Logging in to B again works despite its leftover credentials, and A stays logged out.
	storeTestUser(t)
	if _, err := getOIDCUser(t.Context()); err != nil {
		t.Fatalf("getOIDCUser() after logging in again: %v", err)
	}
	setTestConfigPath(t, dirA)
	if _, err := getOIDCUser(t.Context()); !errors.Is(err, ErrNeedsOIDCLogin) {
		t.Fatalf("getOIDCUser() in logged-out directory = %v, want ErrNeedsOIDCLogin", err)
	}
}

func TestFailedStoreKeepsPreviousCredentials(t *testing.T) {
	skipWithoutUnixPermissions(t)
	setupCredentialStore(t)
	storeTestUser(t)
	key, ciphertext := testEncryptionKey(t), readTestCredentialsFile(t)

	dir := filepath.Dir(*ConfigFilePath)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	tok := (&oauth2.Token{AccessToken: "other"}).WithExtra(map[string]any{"id_token": "other"})
	if _, err := storeOIDCUser(t.Context(), tok, "console.example.com"); err == nil {
		t.Fatal("store into read-only directory succeeded, want error")
	}

	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if testEncryptionKey(t) != key || string(readTestCredentialsFile(t)) != string(ciphertext) {
		t.Fatal("failed store changed the stored credentials")
	}
}

func TestCredentialsFileIsOwnerOnly(t *testing.T) {
	skipWithoutUnixPermissions(t)
	setupCredentialStore(t)

	storeTestUser(t)

	path, err := getCredentialsFilePath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("credentials file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestStoreOIDCUserReplacesMalformedKey(t *testing.T) {
	for _, malformed := range []string{"%%%not-base64", base64.StdEncoding.EncodeToString([]byte("too-short"))} {
		t.Run(malformed, func(t *testing.T) {
			setupCredentialStore(t)
			if err := keyring.Set(malformed); err != nil {
				t.Fatal(err)
			}

			storeTestUser(t)

			if _, err := getOIDCUser(t.Context()); err != nil {
				t.Fatalf("login did not recover from malformed key: %v", err)
			}
		})
	}
}
