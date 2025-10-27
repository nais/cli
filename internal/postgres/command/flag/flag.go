package flag

import (
	"github.com/nais/naistrix"
)

type Postgres struct {
	*naistrix.GlobalFlags
	Namespace Namespace `name:"namespace" short:"n" usage:"The kubernetes |NAMESPACE| to use. Defaults to current namespace."`
	Context   Context   `name:"context" short:"c" usage:"The kubeconfig |CONTEXT| to use. Defaults to current context."`
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

type UserList struct {
	*User
}

type EnableAudit struct {
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
