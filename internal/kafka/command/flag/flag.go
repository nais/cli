package flag

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Kafka struct {
	*flags.GlobalFlags
	Environment Environments `name:"environment" short:"e" usage:"Filter by environment."`
}

type Environments []string

func (e *Environments) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	team := kafkaTeamFromFlags(flags)
	if cliTeam := teamFromCLIArgs(os.Args); cliTeam != "" {
		team = cliTeam
	}
	if team != "" {
		envs, err := kafka.TeamTopicEnvironments(ctx, team)
		if err == nil && len(envs) > 0 {
			return envs, "Available environments"
		}
	}

	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
}

func kafkaTeamFromFlags(flags any) string {
	switch f := flags.(type) {
	case *List:
		return string(f.Team)
	case *Credentials:
		return string(f.Team)
	case *Kafka:
		return string(f.Team)
	default:
		return ""
	}
}

func teamFromCLIArgs(argv []string) string {
	for i := range argv {
		arg := argv[i]

		if after, ok := strings.CutPrefix(arg, "--team="); ok {
			return after
		}
		if after, ok := strings.CutPrefix(arg, "-t="); ok {
			return after
		}

		if arg == "-t" || arg == "--team" {
			if i+1 < len(argv) {
				return argv[i+1]
			}
			return ""
		}
	}

	return ""
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Kafka
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type CredentialsOutput string

var _ naistrix.FlagAutoCompleter = (*CredentialsOutput)(nil)

func (o *CredentialsOutput) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"env", "kcat", "java"}, "Available output formats."
}

type Credentials struct {
	*Kafka
	TTL    string            `name:"ttl" usage:"Time-to-live for the credentials (e.g. '1d', '7d'). Maximum 30 days."`
	Output CredentialsOutput `name:"output" short:"o" usage:"Output format (env, kcat, java). Defaults to env."`
}
