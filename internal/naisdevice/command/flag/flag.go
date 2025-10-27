package flag

import (
	"context"

	"github.com/nais/naistrix"
)

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"yaml", "json"}, "Available output formats."
}

type Status struct {
	*naistrix.GlobalFlags
	Quiet  bool   `name:"quiet" short:"q" usage:"Suppress output"`
	Output Output `name:"output" short:"o" usage:"Format output (yaml|json)."`
}
