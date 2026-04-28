package flag

import (
	"context"
	"fmt"
	"time"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type App struct {
	*flags.GlobalFlags
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
	Output      Output       `name:"output" short:"o" usage:"Format output (table or json)."`
}

func (a *App) GetTeam() string { return a.Team }

type teamScoped interface {
	GetTeam() string
}
type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type instances []string

func (i *instances) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	if args.Len() == 0 {
		return nil, "Please provide an application name to auto-complete instances."
	}

	f, ok := flags.(*Log)
	if !ok {
		return nil, ""
	}
	if len(f.Environment) == 0 {
		return nil, "Please provide environment (-e, --environment) to auto-complete instances."
	}

	if len(f.Team) == 0 {
		return nil, "Please provide team to auto-complete instances. 'nais defaults set team <team>', or '--team <team>' flag."
	}

	instances, err := app.GetApplicationInstances(ctx, f.Team, args.Get("name"), string(f.Environment))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch instances for auto-completion: %v", err)
	}
	return instances, "Available instances"
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Restart struct {
	*App
	Environment Env `name:"environment" short:"e" usage:"Environment of the application. Auto-selected if the app exists in only one environment."`
}

type Issues struct {
	*App
}
type List struct {
	*App
}

type Activity struct {
	*App
	Limit        int           `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
	ActivityType ActivityTypes `name:"activity-type" usage:"Filter by activity type. Can be repeated."`
}

type ActivityTypes []string

func (a *ActivityTypes) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return activityutil.EnumStrings(gql.AllActivityLogActivityType), "Available activity types"
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	if args.Len() == 0 {
		return autoCompleteEnvironments(ctx)
	}

	ts, ok := flags.(teamScoped)
	if !ok {
		return nil, ""
	}
	team := ts.GetTeam()
	if len(team) == 0 {
		return nil, "Please provide team to auto-complete environments. 'nais defaults set team <team>', or '--team <team>' flag."
	}

	envs, err := app.ApplicationEnvironments(ctx, team, args.Get("name"))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Log struct {
	*App
	Environment    Env           `name:"environment" short:"e" usage:"Environment of the application. Auto-selected if the app exists in only one environment."`
	Instance       instances     `name:"instance" short:"i" usage:"Filter by instance. Can be repeated"`
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch. Examples: 300s, 1h, 2h45m. Defaults to 1h."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}

type Status struct {
	*App
	Environment Env `name:"environment" short:"e" usage:"Environment of the application. Auto-selected if the app exists in only one environment."`
}

type EnvVars struct {
	*App
	Environment Env `name:"environment" short:"e" usage:"Environment of the application. Auto-selected if the app exists in only one environment."`
}

type Files struct {
	*App
	Environment Env `name:"environment" short:"e" usage:"Environment of the application. Auto-selected if the app exists in only one environment."`
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}
