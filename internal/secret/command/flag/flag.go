package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Secret struct {
	*flags.GlobalFlags
	Environment Env `name:"environment" short:"e" usage:"Filter by environment."`
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, _ any) ([]string, string) {
	return autoCompleteEnvironments(ctx)
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, _ any) ([]string, string) {
	return autoCompleteEnvironments(ctx)
}

type Output string

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Secret
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
	InUse       bool         `name:"in-use" usage:"Only show secrets that are in use by workloads."`
	Output      Output       `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Get struct {
	*Secret
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Create struct {
	*Secret
}

type Delete struct {
	*Secret
	Yes bool `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type Set struct {
	*Secret
	Key            string `name:"key" usage:"Name of the key to set."`
	Value          string `name:"value" usage:"Value to set."`
	ValueFromStdin bool   `name:"value-from-stdin" usage:"Read value from stdin."`
}

type Unset struct {
	*Secret
	Key string `name:"key" usage:"Name of the key to unset."`
	Yes bool   `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type ViewValues struct {
	*Secret
	Reason string `name:"reason" usage:"Reason for accessing secret values (min 10 chars)."`
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
