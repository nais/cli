package auth

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/nais/cli/internal/keyring"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/oauth2"
)

// ConfigFilePath is a pointer to the global config file path flag.
// It must be set during application initialization (e.g. in application.go)
// so that credentials are stored in the same directory as the config file.
var ConfigFilePath *string

const (
	oidcRefreshTimeout = 20 * time.Second

	// Must outlast a lock holder waiting on keychain prompts, which can take up to 60 seconds per keyring call.
	credentialsLockTimeout    = 5 * time.Minute
	credentialsLockRetryDelay = 50 * time.Millisecond
)

// lockedCredentials accesses the stored credentials while holding a lock shared by all CLI processes.
// The lock is per user rather than per config directory, because all config directories share one keyring key.
type lockedCredentials struct {
	lock *flock.Flock
}

func lockCredentials(ctx context.Context) (*lockedCredentials, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get user config directory: %w", err)
	}

	// The lock file is never deleted, as that would let two processes lock different files.
	lockPath := filepath.Join(userConfigDir, "nais", "credentials.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return nil, fmt.Errorf("create credentials lock directory: %w", err)
	}

	lockCtx, cancel := context.WithTimeout(ctx, credentialsLockTimeout)
	defer cancel()

	lock := flock.New(lockPath, flock.SetPermissions(0o600))
	locked, err := lock.TryLockContext(lockCtx, credentialsLockRetryDelay)
	if err != nil {
		return nil, fmt.Errorf("acquire credentials lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("acquire credentials lock: not acquired")
	}

	return &lockedCredentials{lock: lock}, nil
}

func (c *lockedCredentials) unlock() {
	_ = c.lock.Unlock()
}

// read loads the stored session. Missing or unreadable credentials return ErrNeedsOIDCLogin;
// keyring failures are returned as is.
func (c *lockedCredentials) read() (*oidcUser, error) {
	encryptionKey, err := keyring.GetBytes()
	if err != nil {
		if errors.Is(err, keyring.ErrSecretNotFound) {
			return nil, fmt.Errorf("%w: encryption key not found", ErrNeedsOIDCLogin)
		}
		if errors.Is(err, keyring.ErrInvalidData) {
			return nil, fmt.Errorf("%w: %w", ErrNeedsOIDCLogin, err)
		}
		return nil, fmt.Errorf("get encryption key: %w", err)
	}

	credentialsPath, err := getCredentialsFilePath()
	if err != nil {
		return nil, err
	}

	ciphertext, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("%w: read credentials file: %w", ErrNeedsOIDCLogin, err)
	}

	plaintext, err := decryptCredentials(ciphertext, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt credentials: %w", ErrNeedsOIDCLogin, err)
	}

	var user oidcUser
	if err := json.Unmarshal(plaintext, &user); err != nil {
		return nil, fmt.Errorf("unmarshal oidc user: %w", err)
	}

	return &user, nil
}

// refresh exchanges the refresh token for new tokens and stores them.
func (c *lockedCredentials) refresh(ctx context.Context, user *oidcUser) (*oidcUser, error) {
	ctx, cancel := context.WithTimeout(ctx, oidcRefreshTimeout)
	defer cancel()

	client, err := newOidcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oauth config: %w", err)
	}

	tok, err := client.oauth2.TokenSource(ctx, &user.Token).Token()
	if err != nil {
		return nil, fmt.Errorf("%w: refresh oidc token: %w", ErrNeedsOIDCLogin, err)
	}

	refreshed, err := c.store(tok, user.ConsoleHost)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNeedsOIDCLogin, err)
	}

	return refreshed, nil
}

func (c *lockedCredentials) store(tok *oauth2.Token, consoleURL string) (*oidcUser, error) {
	idToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token")
	}

	user := &oidcUser{
		Token:       *tok,
		IDToken:     idToken,
		ConsoleHost: consoleURL,
	}

	plaintext, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("marshal oidc user: %w", err)
	}

	credentialsPath, err := getCredentialsFilePath()
	if err != nil {
		return nil, fmt.Errorf("get credentials file path: %w", err)
	}

	// Create the directory first, so a failure leaves the keyring untouched.
	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0o700); err != nil {
		return nil, fmt.Errorf("create credentials directory: %w", err)
	}

	encryptionKey, err := c.loadOrCreateEncryptionKey()
	if err != nil {
		return nil, err
	}

	ciphertext, err := encryptCredentials(plaintext, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}

	if err := writeFileAtomic(credentialsPath, ciphertext); err != nil {
		return nil, err
	}

	return user, nil
}

func (c *lockedCredentials) delete() error {
	credentialsPath, err := getCredentialsFilePath()
	if err != nil {
		return err
	}

	// Delete the key first, so a partial failure still logs out every config directory.
	if err := keyring.Delete(); err != nil && !errors.Is(err, keyring.ErrSecretNotFound) {
		return fmt.Errorf("delete encryption key: %w", err)
	}

	if err := os.Remove(credentialsPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete credentials file: %w", err)
	}

	return nil
}

// loadOrCreateEncryptionKey replaces the key only if it's missing or malformed:
// replacing one that merely failed to load would lock out other config directories.
func (c *lockedCredentials) loadOrCreateEncryptionKey() ([]byte, error) {
	encryptionKey, err := keyring.GetBytes()
	if err == nil && len(encryptionKey) == chacha20poly1305.KeySize {
		return encryptionKey, nil
	}
	if err != nil && !errors.Is(err, keyring.ErrSecretNotFound) && !errors.Is(err, keyring.ErrInvalidData) {
		return nil, fmt.Errorf("get encryption key: %w", err)
	}

	encryptionKey = make([]byte, chacha20poly1305.KeySize)
	if _, err := cryptorand.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}

	if err := keyring.SetBytes(encryptionKey); err != nil {
		return nil, fmt.Errorf("set keyring secret: %w", err)
	}

	return encryptionKey, nil
}

func writeFileAtomic(path string, data []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary credentials file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		return fmt.Errorf("write temporary credentials file: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("sync temporary credentials file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temporary credentials file: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace credentials file: %w", err)
	}

	return nil
}

func getCredentialsFilePath() (string, error) {
	const credentialsFileName = "nais-credentials.json.enc"

	if ConfigFilePath == nil || *ConfigFilePath == "" {
		return "", fmt.Errorf("config directory is not set")
	}

	return filepath.Join(filepath.Dir(*ConfigFilePath), credentialsFileName), nil
}

func encryptCredentials(plaintext []byte, encryptionKey []byte) ([]byte, error) {
	const maxPlaintextSize = 64 * 1024 * 1024 // 64 MiB ought to be enough for anyone

	aead, err := chacha20poly1305.NewX(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create aead: %w", err)
	}

	plaintextSize := len(plaintext)
	if plaintextSize > maxPlaintextSize {
		return nil, fmt.Errorf("plaintext too large (%d > %d)", plaintextSize, maxPlaintextSize)
	}

	// Select a random nonce, and leave capacity for the ciphertext.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+plaintextSize+aead.Overhead())
	_, err = cryptorand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Append the ciphertext to the nonce slice so that the output contains both.
	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptCredentials(ciphertext []byte, encryptionKey []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create aead: %w", err)
	}

	if len(ciphertext) < aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext is too short")
	}

	nonce, encrypted := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
	return aead.Open(nil, nonce, encrypted, nil)
}
