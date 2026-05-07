package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Issues struct {
	*flags.GlobalFlags
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}

type (
	IssueType    string
	ResourceName string
	ResourceType string
	Severity     string
)

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Issues
	IssueType    IssueType    `name:"issuetype" usage:"Filter issues by issue type"`
	ResourceName ResourceName `name:"resourcename" usage:"Filter issues by resource name"`
	ResourceType ResourceType `name:"resourcetype" usage:"Filter issues by resource type"`
	Severity     Severity     `name:"severity" usage:"Filter issues by severity"`
}

func (s *Severity) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return toStrings(gql.AllSeverity), "Available severity levels"
}

func (r *ResourceType) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return toStrings(gql.AllResourceType), "Available resource types"
}

func (i *IssueType) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return toStrings(gql.AllIssueType), "Available issue types"
}

func toStrings[T ~string](in []T) []string {
	ret := make([]string, len(in))
	for i, s := range in {
		ret[i] = string(s)
	}
	return ret
}
