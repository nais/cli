package flag

import (
	"context"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
)

type CloudSQL struct {
	*flags.GlobalFlags
	Reason string `name:"reason" short:"r" usage:"Justification for accessing the database. Required for audit logging."`
}

type Password struct {
	*CloudSQL
}

type PasswordRotate struct {
	*Password
}

type User struct {
	*CloudSQL
}

type UserAdd struct {
	*User
	Privilege string `name:"privilege" usage:"The privilege to grant to the user."`
}

type UserDrop struct {
	*User
}

type UserList struct {
	*User
}

type EnableAudit struct {
	*CloudSQL
}

type VerifyAudit struct {
	*CloudSQL
}

type Prepare struct {
	*CloudSQL
	AllPrivileges bool   `name:"all-privileges" usage:"Grant all privileges on the schema to the recipient."`
	Schema        string `name:"schema" usage:"Schema to grant access to."`
	Group         string `name:"group" usage:"Grant access to this Cloud SQL IAM group instead of cloudsqliamuser."`
}

type Proxy struct {
	*CloudSQL
	Port uint   `name:"port" short:"p" usage:"Port to use for the proxy. Defaults to 5432."`
	Host string `name:"host" short:"H" usage:"Host to proxy to. Defaults to localhost."`
}

type Psql struct {
	*CloudSQL
}

type Grant struct {
	*CloudSQL
}

type Revoke struct {
	*CloudSQL
	Schema string `name:"schema" usage:"The schema to revoke privileges from."`
	Group  string `name:"group" usage:"Revoke access from this Cloud SQL IAM group instead of cloudsqliamuser."`
}

type List struct {
	*CloudSQL
	Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
	Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
}

func (*List) LabelFacetResource() string { return "sqlInstances" }

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
