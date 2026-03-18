package flag

import (
	"context"
	"fmt"
	"os"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/cliflags"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/naistrix"
)

type Secret struct {
	*flags.GlobalFlags
	Environment Env `name:"environment" short:"e" usage:"Filter by environment."`
}

type teamProvider interface {
	GetTeam() string
}

func (s *Secret) GetTeam() string { return string(s.Team) }

type Env string

func (e *Env) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, flags any) ([]string, string) {
	team := secretTeamFromFlags(flags)
	if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
		team = cliTeam
	}
	if team != "" {
		envs, err := secret.TeamSecretEnvironments(ctx, team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments with secrets"
		}
	}

	return autoCompleteEnvironments(ctx)
}

// GetEnv is like Env but provides context-aware autocomplete:
//   - when no secret name argument is provided, environments where the team has
//     at least one secret are suggested (falling back to all platform environments)
//   - when a secret name argument has been provided, only environments where
//     that specific secret exists are suggested.
type GetEnv string

func (e *GetEnv) AutoComplete(ctx context.Context, args *naistrix.Arguments, _ string, flags any) ([]string, string) {
	tp, ok := flags.(teamProvider)
	if !ok {
		return nil, "Please provide team to auto-complete environments. 'nais config set team <team>', or '--team <team>' flag."
	}

	team := tp.GetTeam()
	if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
		team = cliTeam
	}

	if team == "" {
		return nil, "Please provide team to auto-complete environments. 'nais config set team <team>', or '--team <team>' flag."
	}

	if args.Len() == 0 {
		envs, err := secret.TeamSecretEnvironments(ctx, team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments with secrets"
		}
		return autoCompleteEnvironments(ctx)
	}

	envs, err := secret.SecretEnvironments(ctx, team, args.Get("name"))
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

func autoCompleteEnvironments(ctx context.Context) ([]string, string) {
	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, flags any) ([]string, string) {
	team := secretTeamFromFlags(flags)
	if cliTeam := cliflags.FirstFlagValue(os.Args, "-t", "--team"); cliTeam != "" {
		team = cliTeam
	}
	if team != "" {
		envs, err := secret.TeamSecretEnvironments(ctx, team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments with secrets"
		}
	}

	return autoCompleteEnvironments(ctx)
}

func secretTeamFromFlags(flags any) string {
	tp, ok := flags.(teamProvider)
	if !ok {
		return ""
	}
	return tp.GetTeam()
}

type Output string

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Secret
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
	Output      Output       `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Activity struct {
	*Secret
	Environment  Environments  `name:"environment" short:"e" usage:"Filter by environment."`
	Output       Output        `name:"output" short:"o" usage:"Format output (table|json)."`
	Limit        int           `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
	ActivityType ActivityTypes `name:"activity-type" usage:"Filter by activity type. Can be repeated."`
}

type ActivityTypes []string

func (a *ActivityTypes) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return activityutil.EnumStrings(gql.AllActivityLogActivityType), "Available activity types"
}

type Get struct {
	*Secret
	Environment GetEnv `name:"environment" short:"e" usage:"Filter by environment."`
	Output      Output `name:"output" short:"o" usage:"Format output (table|json)."`
	WithValues  bool   `name:"with-values" usage:"Also fetch and display secret values (access is logged)."`
	Reason      string `name:"reason" usage:"Reason for accessing secret values (min 10 chars). Used with --with-values."`
}

type Create struct {
	*Secret
}

type Delete struct {
	*Secret
	Environment GetEnv `name:"environment" short:"e" usage:"Filter by environment."`
	Yes         bool   `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type Set struct {
	*Secret
	Environment    GetEnv `name:"environment" short:"e" usage:"Filter by environment."`
	Key            string `name:"key" usage:"Name of the key to set."`
	Value          string `name:"value" usage:"Value to set."`
	ValueFromStdin bool   `name:"value-from-stdin" usage:"Read value from stdin."`
}

type Unset struct {
	*Secret
	Environment GetEnv `name:"environment" short:"e" usage:"Filter by environment."`
	Key         string `name:"key" usage:"Name of the key to unset."`
	Yes         bool   `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}
