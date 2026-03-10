package flag

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type Postgres struct {
	*flags.GlobalFlags
	Namespace   string      `name:"namespace" short:"n" usage:"REMOVED, see --team."`
	Context     string      `name:"context" short:"c" usage:"REMOVED, see --environment."`
	Environment Environment `name:"environment" short:"e" usage:"The |ENVIRONMENT| to use."`
	Reason      string      `name:"reason" short:"r" usage:"Justification for accessing the database. Required for audit logging."`
}

func (p Postgres) UsesRemovedFlags() error {
	if p.Namespace != "" {
		return fmt.Errorf("the --namespace (-n) flag is replaced with the --team (-t) flag")
	}
	if p.Context != "" {
		return fmt.Errorf("the --context (-c) flag is replaced with the --environment (-e) flag")
	}
	return nil
}

type Migrate struct {
	*Postgres
	DryRun bool `name:"dry-run" usage:"Perform a dry run of the migration without applying changes."`
}

type MigrateSetup struct {
	*Migrate
	Tier           string `name:"tier" usage:"The |TIER| of the new instance."`
	DiskAutoResize bool   `name:"disk-auto-resize" usage:"Enable automatic disk resizing for the new instance."`
	DiskSize       int    `name:"disk-size" usage:"The |DISK_SIZE| of the new instance."`
	InstanceType   string `name:"instance-type" usage:"The |TYPE| of the new instance."`
	NoWait         bool   `name:"no-wait" usage:"Do not wait for the job to complete."`
}

type MigratePromote struct {
	*Migrate
	NoWait bool `name:"no-wait" usage:"Do not wait for the job to complete."`
}

type MigrateFinalize struct {
	*Migrate
}

type MigrateRollback struct {
	*Migrate
}

type Password struct {
	*Postgres
}

type PasswordRotate struct {
	*Password
}

type User struct {
	*Postgres
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
	*Postgres
}

type VerifyAudit struct {
	*Postgres
}

type Grant struct {
	*Postgres
}

type Prepare struct {
	*Postgres
	AllPrivileges bool   `name:"all-privileges" usage:"Grant all privileges on the schema to the current user."`
	Schema        string `name:"schema" usage:"Schema to grant access to."`
}

type Proxy struct {
	*Postgres
	Port uint   `name:"port" short:"p" usage:"Port to use for the proxy. Defaults to 5432."`
	Host string `name:"host" short:"H" usage:"Host to proxy to. Defaults to localhost."`
}

type Psql struct {
	*Postgres
}

type Revoke struct {
	*Postgres
	Schema string `name:"schema" usage:"The schema to revoke privileges from."`
}

type List struct {
	*Postgres
	Output Output `name:"output" short:"o" usage:"Format output (table|json)."`
}

type Output string

var _ naistrix.FlagAutoCompleter = (*Output)(nil)

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}
