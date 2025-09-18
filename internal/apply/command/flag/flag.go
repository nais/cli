package flag

import (
	"github.com/nais/cli/internal/root"
)

type filePath string

func (filePath) FileExtensions() (extensions []string) {
	return []string{"toml"}
}

type Apply struct {
	*root.Flags
	Mixin filePath `name:"mixin" short:"m" usage:"Path to the |FILE| containing mixins."`
	Team  string   `name:"team" short:"t" usage:"|TEAM| that owns the resources."`
}
