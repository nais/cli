package flag

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type Users struct {
	*flags.GlobalFlags
}

type ListUsers struct {
	*Users
	Output  Output  `name:"output" short:"o" usage:"Format output (table or json)."`
	Sources Sources `name:"source" usage:"Filter by user source (TENANT or NAIS). Can be repeated."`
	Admin   *bool   `name:"admin" usage:"List only global admins. Use --admin=false for non-admins or omit to include both."`
}

type Sources []string

func (*Sources) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	values := make([]string, len(gql.AllUserSource))
	for i, source := range gql.AllUserSource {
		values[i] = string(source)
	}
	return values, "Available user sources."
}

func (f *ListUsers) Filter() (*gql.UserFilter, error) {
	if f.Sources == nil && f.Admin == nil {
		return nil, nil
	}
	var sources []gql.UserSource
	if f.Sources != nil {
		sources = make([]gql.UserSource, 0, len(f.Sources))
	}
	for _, value := range f.Sources {
		source := gql.UserSource(strings.ToUpper(strings.TrimSpace(value)))
		if !slices.Contains(gql.AllUserSource, source) {
			return nil, fmt.Errorf("invalid --source %q: must be TENANT or NAIS", value)
		}
		sources = append(sources, source)
	}
	return &gql.UserFilter{Sources: sources, IsAdmin: f.Admin}, nil
}
