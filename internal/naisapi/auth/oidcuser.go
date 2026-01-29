package auth

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nais/cli/internal/keyring"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/oauth2"
)

// oidcUser represents a user's session, authenticated via OpenID Connect.
type oidcUser struct {
	oauth2.Token
	IDToken     string `json:"id_token"`
	ConsoleHost string `json:"console_host"`
}

func (u *oidcUser) Refresh(ctx context.Context) (*oidcUser, error) {
	client, err := newOidcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oauth config: %w", err)
	}

	tok, err := client.oauth2.TokenSource(ctx, &u.Token).Token()
	if err != nil {
		return nil, ErrNeedsOIDCLogin
	}

	user, err := storeOIDCUser(tok, u.ConsoleHost)
	if err != nil {
		return nil, fmt.Errorf("%w: %+v", ErrNeedsOIDCLogin, err)
	}

	return user, nil
}

func getOIDCUser(ctx context.Context) (*oidcUser, error) {
	encryptionKey, err := keyring.GetBytes()
	if err != nil {
		if errors.Is(err, keyring.ErrSecretNotFound) || errors.Is(err, keyring.ErrInvalidData) {
			return nil, ErrNeedsOIDCLogin
		}
		return nil, fmt.Errorf("get encryption key: %w", err)
	}

	plaintext, err := readCredentialsFile(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("%w: read credentials file: %+v", ErrNeedsOIDCLogin, err)
	}

	var user oidcUser
	err = json.Unmarshal(plaintext, &user)
	if err != nil {
		return nil, fmt.Errorf("unmarshal oidc user: %w", err)
	}

	if !user.Valid() {
		return user.Refresh(ctx)
	}

	return &user, nil
}

func storeOIDCUser(tok *oauth2.Token, consoleURL string) (*oidcUser, error) {
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

	encryptionKey := make([]byte, chacha20poly1305.KeySize)
	_, err = cryptorand.Read(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}

	if err := keyring.SetBytes(encryptionKey); err != nil {
		return nil, fmt.Errorf("set keyring secret: %w", err)
	}

	if err := writeCredentialsFile(plaintext, encryptionKey); err != nil {
		return nil, err
	}

	return user, nil
}

func getCredentialsFilePath() (string, error) {
	const (
		naisConfigDir       = "nais"
		credentialsFileName = "nais-credentials.json.enc"
	)

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}

	if runtime.GOOS == "darwin" {
		// Respect XDG spec on macOS as os.UserConfigDir does not.
		if dir, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && dir != "" {
			userConfigDir = dir
		}
	}

	return filepath.Join(userConfigDir, naisConfigDir, credentialsFileName), nil
}

func readCredentialsFile(encryptionKey []byte) ([]byte, error) {
	credentialsPath, err := getCredentialsFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(credentialsPath)
	if err != nil {
		return nil, err
	}

	ciphertext, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}

	plaintext, err := decryptCredentials(ciphertext, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials: %w", err)
	}

	return plaintext, nil
}

func writeCredentialsFile(plaintext, encryptionKey []byte) error {
	credentialsPath, err := getCredentialsFilePath()
	if err != nil {
		return fmt.Errorf("get credentials file path: %w", err)
	}

	ciphertext, err := encryptCredentials(plaintext, encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt credentials: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(credentialsPath), 0o700); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}

	if err := os.WriteFile(credentialsPath, ciphertext, 0o600); err != nil {
		return fmt.Errorf("write credentials file: %w", err)
	}

	return nil
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
