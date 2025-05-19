// Package keyring is a simple wrapper that adds timeouts to the zalando/go-keyring package
// "Borrowed" with ❤️ from https://github.com/cli/cli/blob/17af24e147629aa1aed2546e87e9323aeabf4c8c/internal/keyring/keyring.go
package nais

import (
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "cli.nais.io"
	user        = "nais-user"
)

var errSecretNotFound = errors.New("secret not found in keyring")

type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string {
	return e.message
}

// Set secret in keyring.
func setSecret(secret string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(serviceName, user, secret)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &timeoutError{"timeout while trying to set secret in keyring"}
	}
}

// Get secret from keyring.
func getSecret() (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := keyring.Get(serviceName, user)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()
	select {
	case res := <-ch:
		if errors.Is(res.err, keyring.ErrNotFound) {
			return "", errSecretNotFound
		}
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", &timeoutError{"timeout while trying to get secret from keyring"}
	}
}

// Delete secret from keyring.
func deleteSecret() error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(serviceName, user)
	}()
	select {
	case err := <-ch:
		if errors.Is(err, keyring.ErrNotFound) {
			return errSecretNotFound
		}
		return err
	case <-time.After(3 * time.Second):
		return &timeoutError{"timeout while trying to delete secret from keyring"}
	}
}
