package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Issues struct {
	*flag.Alpha
}

type (
	IssueType    string
	ResourceName string
	ResourceType string
	Severity     string
	Environment  string
)

type List struct {
	*Issues
	Environment  Environment  `name:"environment" usage:"Filter issues by environment"` // TODO: Find and list environments
	IssueType    IssueType    `name:"issuetype" usage:"Filter issues by issue type"`
	ResourceName ResourceName `name:"resourcename" usage:"Filter issues by resource name"` // TODO: Find resource names in current team
	ResourceType ResourceType `name:"resourcetype" usage:"Filter issues by resource type"`
	Severity     Severity     `name:"severity" usage:"Filter issues by severity"`
}

func (e *Environment) AutoComplete(ctx context.Context, args *naistrix.Arguments, str string, flags any) ([]string, string) {
	// f := flags.(*List)
	// fmt.Println("IssueType", f.IssueType)

	envs, err := naisapi.GetAllEnvironments(ctx)
	if err != nil {
		return nil, fmt.Sprintf("Failed to fetch environments for auto-completion: %v", err)
	}
	return envs, "Available environments"
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
