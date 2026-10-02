package cloudsql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/naistrix"
)

var grantAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON TABLES TO $role;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON SEQUENCES TO $role;
	GRANT ALL ON ALL TABLES IN SCHEMA $schema TO $role;
	GRANT ALL ON ALL SEQUENCES IN SCHEMA $schema TO $role;
	GRANT CREATE ON SCHEMA $schema TO $role;`

var grantSelectPrivs = `GRANT USAGE ON SCHEMA $schema TO $role;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON TABLES TO $role;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON SEQUENCES TO $role;
	GRANT SELECT ON ALL TABLES IN SCHEMA $schema TO $role;
	GRANT SELECT ON ALL SEQUENCES IN SCHEMA $schema TO $role;`

// this is used for all privileges and select, as it covers both cases
var revokeAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON TABLES FROM $role;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON SEQUENCES FROM $role;
	REVOKE ALL ON ALL TABLES IN SCHEMA $schema FROM $role;
	REVOKE ALL ON ALL SEQUENCES IN SCHEMA $schema FROM $role;
	REVOKE CREATE ON SCHEMA $schema FROM $role;`

var (
	grantUsage  = `GRANT USAGE ON SCHEMA $schema TO $role;`
	revokeUsage = `REVOKE USAGE ON SCHEMA $schema FROM $role;`
)

func PrepareAccess(ctx context.Context, appName, team, environment string, fl *flag.Prepare, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	sv, err := GetSecretValues(ctx, appName, team, environment, fl.CloudSQL, ReasonPrepareAccess, out)
	if err != nil {
		return err
	}

	prependUsageIfNotPublic := func(statement string) string {
		if fl.Schema != "public" {
			return grantUsage + "\n" + statement
		}
		return statement
	}

	if fl.AllPrivileges {
		return sqlExecAsAppUser(ctx, appName, team, environment, fl.Schema, fl.Group, prependUsageIfNotPublic(grantAllPrivs), sv)
	} else {
		return sqlExecAsAppUser(ctx, appName, team, environment, fl.Schema, fl.Group, prependUsageIfNotPublic(grantSelectPrivs), sv)
	}
}

func RevokeAccess(ctx context.Context, appName, team, environment string, fl *flag.Revoke, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	sv, err := GetSecretValues(ctx, appName, team, environment, fl.CloudSQL, ReasonRevokeAccess, out)
	if err != nil {
		return err
	}

	q := revokeAllPrivs
	// Keep the existing public-schema behavior for individual IAM users; undo the
	// explicit USAGE grant for groups when revoking group access.
	if fl.Schema != "public" || fl.Group != "" {
		q += "\n" + revokeUsage
	}
	return sqlExecAsAppUser(ctx, appName, team, environment, fl.Schema, fl.Group, q, sv)
}

func sqlExecAsAppUser(ctx context.Context, appName, team, environment string, schema, group, statement string, sv *SecretValues) error {
	dbInfo, err := NewDBInfo(ctx, appName, team, environment)
	if err != nil {
		return err
	}

	dbInfo.SetSecretValues(sv)

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if group != "" {
		var isGroup bool
		err := db.QueryRowContext(ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_roles AS account
			JOIN pg_auth_members AS membership ON membership.member = account.oid
			WHERE account.rolname = $1 AND membership.roleid = to_regrole('cloudsqliamgroup')
		)`, group).Scan(&isGroup)
		if err != nil {
			return fmt.Errorf("verify Cloud SQL IAM group: %w", formatInvalidGrantError(err))
		}
		if !isGroup {
			return fmt.Errorf("%q is not a Cloud SQL IAM group on this instance", group)
		}
	}

	_, err = db.ExecContext(ctx, accessStatement(statement, schema, group))
	if err != nil {
		return formatInvalidGrantError(err)
	}

	return nil
}

func accessStatement(statement, schema, group string) string {
	role := "cloudsqliamuser"
	if group != "" {
		role = pq.QuoteIdentifier(group)
	}
	return strings.NewReplacer("$schema", pq.QuoteIdentifier(schema), "$role", role).Replace(statement)
}
