package flag

import (
	"context"
	"slices"
	"strings"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type (
	Output                string
	CredentialsOutput     string
	KafkaTopicGrantAccess string
)

var (
	_ naistrix.FlagAutoCompleter = (*Output)(nil)
	_ naistrix.FlagAutoCompleter = (*CredentialsOutput)(nil)
	_ naistrix.FlagAutoCompleter = (*KafkaTopicGrantAccess)(nil)
)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

func (o *CredentialsOutput) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"env", "kcat", "java"}, "Available output formats."
}

func (a *KafkaTopicGrantAccess) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return KafkaTopicGrantAccessChoices(), "Available access levels."
}

// KafkaTopicGrantAccessChoices returns the GraphQL access levels in their CLI lowercase form.
func KafkaTopicGrantAccessChoices() []string {
	choices := make([]string, len(gql.AllKafkaTopicGrantAccess))
	for i, access := range gql.AllKafkaTopicGrantAccess {
		choices[i] = strings.ToLower(string(access))
	}
	return choices
}

func (a KafkaTopicGrantAccess) Validate() error {
	valid := KafkaTopicGrantAccessChoices()
	if a == "" {
		return naistrix.Errorf("access level is required, must be one of: %s", strings.Join(valid, ", "))
	}

	if !slices.Contains(valid, string(a)) {
		return naistrix.Errorf("invalid access level: %q, must be one of: %s", a, strings.Join(valid, ", "))
	}

	return nil
}

type Kafka struct {
	*flags.GlobalFlags
}

type List struct {
	*Kafka
	Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
	Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
}

func (*List) LabelFacetResource() string { return "kafkaTopics" }

type Credentials struct {
	*Kafka
	TTL    string            `name:"ttl" usage:"Time-to-live for the credentials (e.g. '1d', '7d'). Maximum 30 days."`
	Output CredentialsOutput `name:"output" short:"o" usage:"Output format (env, kcat, java). Defaults to env."`
}

type GrantAccess struct {
	*Kafka
	Access KafkaTopicGrantAccess `name:"access" short:"a" usage:"Access |LEVEL| (readwrite, read and write)."`
}

type ListGrants struct {
	*Kafka
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}
