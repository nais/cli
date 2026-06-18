package flag

import (
	"context"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
)

type Job struct {
	*flags.GlobalFlags
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Job
	Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
	Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
}

func (*List) LabelFacetResource() string { return "jobs" }

type Issues struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}

type Activity struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
	Limit  int    `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
}

type Trigger struct {
	*Job
	RunName string `name:"run-name" usage:"Custom run name. Defaults to a generated value."`
}

type Delete struct {
	*Job
}

type RunList struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}

type Log struct {
	*Job
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch. Examples: 300s, 1h, 2h45m. Defaults to 1h."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}

type SetEnv struct {
	*Job
}
