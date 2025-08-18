package flag

import (
	"github.com/nais/cli/internal/root"
)

type filePath string

func (filePath) FileExtensions() (extensions []string) {
	return []string{"json", "yaml", "yml", "toml"}
}

type Apply struct {
	*root.Flags
	FilePath filePath `name:"file" short:"f" usage:"Path to the |FILE| containing resource definitions."`
}
