// keyring is a simple wrapper that adds timeouts to the zalando/go-keyring package
// "Borrowed" with ❤️ from https://github.com/cli/cli/blob/17af24e147629aa1aed2546e87e9323aeabf4c8c/internal/keyring/keyring.go

package auth

import (
	"errors"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "cli.nais.io"
	keyringUser = "nais-user"
)

var errSecretNotFound = errors.New("secret not found in keyring")

type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string {
	return e.message
}

func setKeyringSecret(secret string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(serviceName, keyringUser, secret)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &timeoutError{"timeout while trying to set secret in keyring"}
	}
}

func getKeyringSecret() (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := keyring.Get(serviceName, keyringUser)
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

func deleteKeyringSecret() error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(serviceName, keyringUser)
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
