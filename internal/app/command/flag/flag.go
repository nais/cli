package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type App struct {
	*flags.GlobalFlags
	Environment []string `name:"environment" short:"e" usage:"Filter by environment."`
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Issues struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
type ListApps struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
