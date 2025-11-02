package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"yaml", "json"}, "Available output formats."
}

type Device struct {
	*flags.GlobalFlags
}

type Status struct {
	*Device
	Quiet  bool   `name:"quiet" short:"q" usage:"Suppress output"`
	Output Output `name:"output" short:"o" usage:"Format output (yaml|json)."`
}

type Gateway struct {
	*Device
}

type List struct {
	*Gateway
	Output Output `name:"output" short:"o" usage:"Format output (yaml|json)."`
}

type Describe struct {
	*Gateway
	Output Output `name:"output" short:"o" usage:"Format output (yaml|json)."`
}

type Connect struct {
	*Gateway
}
