package cli

import (
	"fmt"

	"github.com/pterm/pterm"
)

type Err struct {
	Message string
}

func (e Err) Error() string {
	return pterm.Error.Sprintf("%s", e.Message)
}

func Errorf(format string, a ...any) error {
	return Err{Message: fmt.Sprintf(format, a...)}
}
