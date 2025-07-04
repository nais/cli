package flag

import (
	"context"

	"github.com/nais/cli/internal/root"
)

type Alpha struct {
	*root.Flags
}

type Api struct {
	*Alpha
}

type Proxy struct {
	*Api
	ListenAddr string `name:"listen" short:"l" usage:"Address the proxy will listen on."`
}

type Output string

func (o *Output) AutoComplete(context.Context, []string, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Team struct {
	*Api
}

type AddMember struct {
	*Team
	Owner bool `name:"owner" short:"o" usage:"Assign owner role to the member."`
}

type RemoveMember struct {
	*Team
}

type ListMembers struct {
	*Team
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type ListWorkloads struct {
	*Team
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Teams struct {
	*Api
	All    bool   `name:"all" short:"a" usage:"List all teams, not just the ones you are a member of."`
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Schema struct {
	*Api
}

type Status struct {
	*Api
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}
