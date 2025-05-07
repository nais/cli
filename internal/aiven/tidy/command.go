package tidy

import (
	"github.com/nais/cli/internal/aiven"
)

func Run() error {
	return aiven.TidyLocalSecrets()
}
