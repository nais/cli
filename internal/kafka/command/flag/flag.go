package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
)

type Kafka struct {
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

type List struct {
	*Kafka
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
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

type GrantAccess struct {
	*Kafka
	Environment Environment `name:"environment" short:"e" usage:"The |ENVIRONMENT| to use."`
	Access      string      `name:"access" short:"a" usage:"Access |LEVEL| (readwrite, read and write)."`
}
