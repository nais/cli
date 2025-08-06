package flag

import (
	"github.com/nais/cli/internal/root"
)

type varsFilePath string

func (varsFilePath) FileExtensions() (extensions []string) {
	return []string{"json", "yaml", "yml"}
}

type Validate struct {
	*root.Flags
	VarsFilePath varsFilePath `name:"vars-file" short:"f" usage:"Path to the |FILE| containing template variables in JSON or YAML format."`
	Vars         []string     `name:"var" usage:"Template variable in |KEY=VALUE| form. Can be repeated."`
}
