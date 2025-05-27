package flag

import (
	"github.com/nais/cli/internal/root"
)

type Postgres struct {
	*root.Flags
	Namespace string
	Context   string
}

type Migrate struct {
	*Postgres
	DryRun bool
}

type MigrateSetup struct {
	*Migrate
	Tier           string
	DiskAutoResize bool
	DiskSize       int
	InstanceType   string
	NoWait         bool
}

type MigratePromote struct {
	*Migrate
	NoWait bool
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
	Privilege string
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
	AllPrivileges bool
	Schema        string
}

type Proxy struct {
	*Postgres
	Port uint
	Host string
}

type Psql struct {
	*Postgres
}

type Revoke struct {
	*Postgres
	Schema string
}
