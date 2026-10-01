package flag

import (
	"context"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
)

type (
	Postgres struct{ *flags.GlobalFlags }
	Access   struct {
		*Postgres
		Branch      string        `name:"branch" usage:"Branch to access (defaults to the active branch)."`
		AccessLevel string        `name:"access-level" usage:"Access level: read, write, or admin."`
		Reason      string        `name:"reason" usage:"Reason for personal access (at least 10 characters)."`
		TTL         time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Database    string        `name:"database" usage:"Database name for psql (default app)."`
	}
	Proxy struct {
		*Postgres
		Branch        string        `name:"branch" usage:"Branch to access (defaults to the active branch)."`
		AccessLevel   string        `name:"access-level" usage:"Access level: read, write, or admin."`
		Reason        string        `name:"reason" usage:"Reason for personal access (at least 10 characters)."`
		TTL           time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Host          string        `name:"host" usage:"Local loopback address (default 127.0.0.1)."`
		Port          int           `name:"port" usage:"Local port (default random)."`
		PrintPassword bool          `name:"print-password" usage:"Print the database password to stdout (sensitive)."`
	}
	List struct {
		*Postgres
		Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
		Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
	}
)

func (*List) LabelFacetResource() string { return "postgresBranches" }

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
