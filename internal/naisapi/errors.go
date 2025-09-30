package naisapi

import (
	"errors"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

func IsErrAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	var gerr gqlerror.List
	ok := errors.As(err, &gerr)
	if !ok {
		return false
	}

	for _, e := range gerr {
		if e.Message == "Resource already exists." {
			return true
		}
	}

	return false
}
