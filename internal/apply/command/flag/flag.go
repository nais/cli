package flag

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/naistrix"
)

type filePath string

var _ naistrix.FileAutoCompleter = (*filePath)(nil)

func (filePath) FileExtensions() (extensions []string) {
	return []string{"toml"}
}

type Apply struct {
	*alpha.Alpha
	Mixin filePath `name:"mixin" short:"m" usage:"Path to the |FILE| containing mixins."`
}
