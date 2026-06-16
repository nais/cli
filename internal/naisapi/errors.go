package naisapi

import "strings"

func IsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Resource not found:")
}
