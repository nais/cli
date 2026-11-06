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
		Branch      string        `name:"branch" usage:"Deprecated: omit this flag. Only the legacy value main is accepted."`
		AccessLevel AccessLevel   `name:"access-level" usage:"Access level: read, write, admin (application database owner privileges), or maintain (instance monitoring and maintenance)."`
		Reason      string        `name:"reason" usage:"Reason for personal access (at least 10 characters); prompted if omitted interactively."`
		TTL         time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Database    string        `name:"database" usage:"Database name for psql (default app)."`
	}
	Proxy struct {
		*Postgres
		Branch         string        `name:"branch" usage:"Deprecated: omit this flag. Only the legacy value main is accepted."`
		AccessLevel    AccessLevel   `name:"access-level" usage:"Access level: read, write, admin (application database owner privileges), or maintain (instance monitoring and maintenance)."`
		Reason         string        `name:"reason" usage:"Reason for personal access (at least 10 characters); prompted if omitted interactively."`
		TTL            time.Duration `name:"ttl" usage:"Requested lifetime (default 30m, maximum 1h)."`
		Host           string        `name:"host" usage:"Local loopback address (default 127.0.0.1)."`
		Port           int           `name:"port" short:"p" usage:"Local port (default random)."`
		Output         ProxyOutput   `name:"output" short:"o" usage:"Format output (human, env, or json)."`
		PrintPassword  bool          `name:"print-password" usage:"Print the database password to stdout (sensitive)."`
		ExitOnStdinEOF bool          `name:"exit-on-stdin-eof" usage:"Stop the proxy when stdin reaches EOF after setup (disabled by default)."`
	}
	InstanceCreate struct {
		*Postgres
		Yes              bool            `name:"yes" short:"y" usage:"Confirm without a prompt."`
		Version          PostgresVersion `name:"version" usage:"PostgreSQL major version (18)."`
		HighAvailability bool            `name:"high-availability" usage:"Enable high availability (third instance and synchronous replication)."`
		CPU              string          `name:"cpu" usage:"Requested CPU, e.g. 100m."`
		Memory           string          `name:"memory" usage:"Requested memory, e.g. 512Mi."`
		DiskSize         string          `name:"disk-size" usage:"Requested disk size, e.g. 10Gi."`
	}
	InstanceDelete struct {
		*Postgres
		Yes bool `name:"yes" short:"y" usage:"Confirm irreversible deletion without a prompt."`
	}
	InstanceUpdate struct {
		*Postgres
		Yes              bool             `name:"yes" short:"y" usage:"Confirm without a prompt."`
		HighAvailability HighAvailability `name:"high-availability" usage:"Enable or disable high availability (true or false)."`
		CPU              string           `name:"cpu" usage:"Requested CPU, e.g. 100m."`
		Memory           string           `name:"memory" usage:"Requested memory, e.g. 512Mi."`
		DiskSize         string           `name:"disk-size" usage:"Requested disk size, e.g. 10Gi."`
	}
	BranchList struct {
		*Postgres
		Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
	}
	List struct {
		*Postgres
		Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
		Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
	}
)

func (*List) LabelFacetResource() string { return "postgreses" }

// AccessLevel completes the personal database and instance access levels.
type AccessLevel string

var _ naistrix.FlagAutoCompleter = (*AccessLevel)(nil)

func (*AccessLevel) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"read", "write", "admin", "maintain"}, "Personal access levels; admin grants application database owner privileges, maintain grants instance monitoring and maintenance."
}

// PostgresVersion completes supported major versions for creation.
type PostgresVersion string

var _ naistrix.FlagAutoCompleter = (*PostgresVersion)(nil)

func (*PostgresVersion) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"18"}, "Supported PostgreSQL versions."
}

// HighAvailability completes explicit update values, including false.
type HighAvailability string

var _ naistrix.FlagAutoCompleter = (*HighAvailability)(nil)

func (*HighAvailability) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"true", "false"}, "High availability values."
}

type Output string

// ProxyOutput completes output formats for the long-running proxy command.
type ProxyOutput string

var _ naistrix.FlagAutoCompleter = (*ProxyOutput)(nil)

func (*ProxyOutput) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"human", "env", "json"}, "Available proxy output formats."
}

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
