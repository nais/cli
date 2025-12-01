package flag

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type App struct {
	*flags.GlobalFlags
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
}
type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Restart struct {
	*App
}

type Issues struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
type List struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	//TODO: only return environments with Applications
	return autoCompleteEnvironments(ctx)
}

type Log struct {
	*App
	Environment    Env           `name:"environment" short:"e" usage:"Filter by environment."`
	Team           []string      `name:"team" short:"t" usage:"Filter logs to a single |team|. Can be repeated."`
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}
