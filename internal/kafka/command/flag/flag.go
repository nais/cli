package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type Kafka struct {
	*flags.GlobalFlags
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
	Access string `name:"access" short:"a" usage:"Access |LEVEL| (readwrite, read and write)."`
}
