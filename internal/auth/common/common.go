package common

import (
	"fmt"

	"github.com/nais/naistrix"
)

func Deprecated(cmd *naistrix.Command) {
	cmd.Title = fmt.Sprintf("%s (Deprecated, use `nais auth %s`)", cmd.Title, cmd.Name)
}
