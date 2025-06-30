package flag

import (
	"context"

	"github.com/nais/cli/v2/internal/root"
)

type Output string

type Status struct {
	*root.Flags
	Quiet  bool   `name:"quiet" short:"q" usage:"Suppress output"`
	Output Output `name:"output" short:"o" usage:"Format output (yaml|json)."`
}

func (o *Output) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	return []string{"yaml", "json"}, "Available output formats."
}
