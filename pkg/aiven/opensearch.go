package aiven

import (
	"fmt"
	"strings"
)

type OpenSearchAccess int64

const (
	Read OpenSearchAccess = iota
	Write
	ReadWrite
	Admin
)

var OpenSearchAccesses = []string{"read", "write", "readwrite", "admin"}

func OpenSearchAccessFromString(access string) (OpenSearchAccess, error) {
	switch strings.ToLower(access) {
	case OpenSearchAccesses[0]:
		return Read, nil
	case OpenSearchAccesses[1]:
		return Write, nil
	case OpenSearchAccesses[2]:
		return ReadWrite, nil
	case OpenSearchAccesses[3]:
		return Admin, nil
	default:
		return -1, fmt.Errorf("unknown access: %v", access)
	}
}

func (p OpenSearchAccess) String() string {
	return OpenSearchAccesses[p]
}
