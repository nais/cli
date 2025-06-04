package flag

import (
	"github.com/nais/cli/internal/root"
)

type Validate struct {
	*root.Flags
	VarsFilePath string   `name:"vars-file" short:"f" usage:"Path to the FILE containing template variables in JSON or YAML format."`
	Vars         []string "name:\"var\" short:\"v\" usage:\"Template variable in `KEY=VALUE` form. Can be repeated.\""
}
