package flag

import (
	"github.com/nais/cli/internal/root"
)

type Validate struct {
	*root.Flags
	VarsFilePath string
	Vars         []string
}
