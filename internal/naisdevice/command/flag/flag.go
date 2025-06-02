package flag

import "github.com/nais/cli/internal/root"

type Status struct {
	*root.Flags
	Quiet  bool   `name:"quiet" short:"q" usage:"Suppress output"`
	Output string `name:"output" short:"o" usage:"Output format (json, yaml)"`
}
