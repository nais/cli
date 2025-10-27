package flag

import (
	"github.com/nais/naistrix"
)

type varsFilePath string

var _ naistrix.FileAutoCompleter = (*varsFilePath)(nil)

func (varsFilePath) FileExtensions() (extensions []string) {
	return []string{"json", "yaml", "yml"}
}

type Validate struct {
	*naistrix.GlobalFlags
	VarsFilePath varsFilePath `name:"vars-file" short:"f" usage:"Path to the |FILE| containing template variables in JSON or YAML format."`
	Vars         []string     `name:"var" usage:"Template variable in |KEY=VALUE| form. Can be repeated."`
}
