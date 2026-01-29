// keyring is a simple wrapper that adds timeouts to the zalando/go-keyring package
// "Borrowed" with ❤️ from https://github.com/cli/cli/blob/17af24e147629aa1aed2546e87e9323aeabf4c8c/internal/keyring/keyring.go

package keyring

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "cli.nais.io"
	keyringUser = "nais-user"
)

var (
	ErrSecretNotFound = errors.New("secret not found in keyring")
	ErrInvalidData    = errors.New("invalid data stored in keyring")
)

type TimeoutError struct {
	message string
}

func (e *TimeoutError) Error() string {
	return e.message
}

func Get() (string, error) {
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
			return "", ErrSecretNotFound
		}
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", &TimeoutError{"timeout while trying to get secret from keyring"}
	}
}

func GetBytes() ([]byte, error) {
	encoded, err := Get()
	if err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: decode base64: %+v", ErrInvalidData, err)
	}
	return bytes, nil
}

func Set(value string) error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Set(serviceName, keyringUser, value)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to set secret in keyring"}
	}
}

func SetBytes(value []byte) error {
	return Set(base64.StdEncoding.EncodeToString(value))
}

func Delete() error {
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		ch <- keyring.Delete(serviceName, keyringUser)
	}()
	select {
	case err := <-ch:
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrSecretNotFound
		}
		return err
	case <-time.After(3 * time.Second):
		return &TimeoutError{"timeout while trying to delete keyring secrets for service"}
	}
}
