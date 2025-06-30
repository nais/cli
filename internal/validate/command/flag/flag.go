package flag

import (
	"github.com/nais/cli/v2/internal/root"
)

type Validate struct {
	*root.Flags
	VarsFilePath string   `name:"vars-file" short:"f" usage:"Path to the |FILE| containing template variables in JSON or YAML format."`
	Vars         []string `name:"var" usage:"Template variable in |KEY=VALUE| form. Can be repeated."`
}
