package flag

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type App struct {
	*flags.GlobalFlags
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
}
type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	team := appTeamFromFlags(flags)
	if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
		team = cliTeam
	}
	if team == "" {
		return nil, "Please provide team to auto-complete environments. 'nais config set team <team>', or '--team <team>' flag."
	}

	if appName := appNameForEnvironmentCompletion(args); appName != "" {
		envs, err := app.ApplicationEnvironments(ctx, team, appName)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments"
		}
	}

	envs, err := app.TeamApplicationEnvironments(ctx, team)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	if len(envs) > 0 {
		return envs, "Available environments"
	}

	return nil, "No environments with applications found for this team."
}

func appTeamFromFlags(flags any) string {
	switch f := flags.(type) {
	case *Activity:
		if f.App != nil && f.App.Team != "" {
			return string(f.App.Team)
		}
		return string(f.Team)
	case *Issues:
		if f.App != nil && f.App.Team != "" {
			return string(f.App.Team)
		}
		return string(f.Team)
	case *List:
		if f.App != nil && f.App.Team != "" {
			return string(f.App.Team)
		}
		return string(f.Team)
	case *Restart:
		return string(f.Team)
	case *App:
		return string(f.Team)
	default:
		v := reflect.ValueOf(flags)
		if !v.IsValid() {
			return ""
		}
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return ""
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return ""
		}

		teamField := v.FieldByName("Team")
		if teamField.IsValid() && teamField.Kind() == reflect.String {
			return teamField.String()
		}

		return ""
	}
}

func appNameForEnvironmentCompletion(args *naistrix.Arguments) string {
	if args.Len() > 0 {
		if name := args.Get("name"); name != "" {
			return name
		}
	}

	// Some command/flag combinations (for example app restart with parent sticky flags)
	// do not expose positional args through naistrix during completion.
	if !isRestartCompletionFromCLIArgs() {
		return ""
	}

	return appNameFromCLIArgs(os.Args)
}

func isRestartCompletionFromCLIArgs() bool {
	return cliflags.HasSubCommandPath(os.Args, "app", "restart")
}

func appNameFromCLIArgs(argv []string) string {
	seenRestart := false

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		if arg == "restart" {
			seenRestart = true
			continue
		}
		if !seenRestart {
			continue
		}

		if arg == "--" {
			if i+1 < len(argv) {
				return argv[i+1]
			}
			return ""
		}

		if strings.HasPrefix(arg, "--team=") || strings.HasPrefix(arg, "--environment=") || strings.HasPrefix(arg, "--config=") {
			continue
		}

		if arg == "-t" || arg == "--team" || arg == "-e" || arg == "--environment" || arg == "--config" {
			i++
			continue
		}

		if strings.HasPrefix(arg, "-") {
			continue
		}

		return arg
	}

	return ""
}

type instances []string

func (i *instances) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	if args.Len() == 0 {
		return nil, "Please provide an application name to auto-complete instances."
	}

	f := flags.(*Log)
	if len(f.Environment) == 0 {
		return nil, "Please provide environment (--environment/-e) to auto-complete instances."
	}

	if len(f.Team) == 0 {
		return nil, "Please provide team to auto-complete instances. 'nais config team set <team>', or '--team <team>' flag."
	}

	instances, err := app.GetApplicationInstances(ctx, string(f.Team), args.Get("name"), string(f.Environment))
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
}

type Issues struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
type List struct {
	*App
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Activity struct {
	*App
	Output       Output        `name:"output" short:"o" usage:"Format output (table|json)."`
	Limit        int           `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
	ActivityType ActivityTypes `name:"activity-type" usage:"Filter by activity type. Can be repeated."`
}

type ActivityTypes []string

func (a *ActivityTypes) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return activityutil.EnumStrings(gql.AllActivityLogActivityType), "Available activity types"
}

type Env string

func (e *Env) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	f := flags.(*Log)
	if len(f.Team) == 0 {
		return nil, "Please provide team to auto-complete environments. 'nais config team set <team>', or '--team <team>' flag."
	}

	if args.Len() == 0 {
		envs, err := app.TeamApplicationEnvironments(ctx, f.Team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments"
		}
		return autoCompleteEnvironments(ctx)
	}

	envs, err := app.ApplicationEnvironments(ctx, f.Team, args.Get("name"))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Log struct {
	*App
	Environment    Env           `name:"environment" short:"e" usage:"Filter by environment."`
	Instance       instances     `name:"instance" short:"i" usage:"Filter by instance. Can be repeated"`
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch. Examples: 300s, 1h, 2h45m. Defaults to 1h."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}
