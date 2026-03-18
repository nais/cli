package flag

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Job struct {
	*flags.GlobalFlags
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
}

func (j *Job) GetTeam() string {
	if j == nil {
		return ""
	}
	return string(j.Team)
}

type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	team := jobTeamFromFlags(flags)
	if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
		team = cliTeam
	}
	if team == "" {
		envs, err := naisapi.GetAllEnvironments(ctx)
		if err != nil {
			return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
		}
		return envs, "Available environments"
	}

	if jobName := jobNameForEnvironmentCompletion(args); jobName != "" {
		envs, err := job.JobEnvironments(ctx, team, jobName)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments"
		}
	}

	envs, err := job.TeamJobEnvironments(ctx, team)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	if len(envs) > 0 {
		return envs, "Available environments"
	}

	return nil, "No environments with jobs found for this team."
}

type teamProvider interface {
	GetTeam() string
}

func jobTeamFromFlags(flags any) string {
	tp, ok := flags.(teamProvider)
	if !ok {
		return ""
	}
	return tp.GetTeam()
}

func jobNameForEnvironmentCompletion(args *naistrix.Arguments) string {
	if args.Len() > 0 {
		if name := args.Get("name"); name != "" {
			return name
		}
	}

	if !isTriggerCompletionFromCLIArgs() {
		return ""
	}

	return jobNameFromCLIArgs(os.Args)
}

func isTriggerCompletionFromCLIArgs() bool {
	return cliflags.HasSubCommandPathWithValueFlags(
		os.Args,
		"job",
		[]string{"-t", "--team", "-e", "--environment", "--config", "--run-name"},
		"trigger",
	)
}

func jobNameFromCLIArgs(argv []string) string {
	return cliflags.PositionalArgAfterSubcommand(
		argv,
		"trigger",
		[]string{"-t", "--team", "-e", "--environment", "--config", "--run-name"},
	)
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Issues struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Activity struct {
	*Job
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
	Limit  int    `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
}

type Trigger struct {
	*Job
	Environment Env    `name:"environment" short:"e" usage:"Filter by environment."`
	RunName     string `name:"run-name" usage:"Custom run name. Defaults to a generated value."`
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	var team string
	switch t := flags.(type) {
	case *Log:
		team = string(t.Team)
	case *Trigger:
		team = string(t.Team)
	}
	if team == "" {
		if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
			team = cliTeam
		}
	}
	if team == "" {
		return nil, "Please provide team to auto-complete environments. 'nais config set team <team>', or '--team <team>' flag."
	}

	if args.Len() == 0 {
		envs, err := job.TeamJobEnvironments(ctx, team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments"
		}
		envs, err = naisapi.GetAllEnvironments(ctx)
		if err != nil {
			return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
		}
		return envs, "Available environments"
	}

	envs, err := job.JobEnvironments(ctx, team, args.Get("name"))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Log struct {
	*Job
	Environment    Env           `name:"environment" short:"e" usage:"Filter by environment."`
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch. Examples: 300s, 1h, 2h45m. Defaults to 1h."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}
