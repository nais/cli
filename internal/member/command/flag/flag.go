package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type Member struct {
	*flags.GlobalFlags
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type AddMember struct {
	*Member
	Owner bool `name:"owner" short:"o" usage:"Assign owner role to the member."`
}

type SetRole struct {
	*Member
}

type RemoveMember struct {
	*Member
}

type ListMembers struct {
	*Member
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
