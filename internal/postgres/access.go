package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

var grantAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON SEQUENCES TO cloudsqliamuser;
	GRANT ALL ON ALL TABLES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT ALL ON ALL SEQUENCES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT CREATE ON SCHEMA $schema TO cloudsqliamuser;`

var grantSelectPrivs = `GRANT USAGE ON SCHEMA $schema TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON SEQUENCES TO cloudsqliamuser;
	GRANT SELECT ON ALL TABLES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT SELECT ON ALL SEQUENCES IN SCHEMA $schema TO cloudsqliamuser;`

// this is used for all privileges and select, as it covers both cases
var revokeAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON TABLES FROM cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON SEQUENCES FROM cloudsqliamuser;
	REVOKE ALL ON ALL TABLES IN SCHEMA $schema FROM cloudsqliamuser;
	REVOKE ALL ON ALL SEQUENCES IN SCHEMA $schema FROM cloudsqliamuser;
	REVOKE CREATE ON SCHEMA $schema FROM cloudsqliamuser;`

var (
	grantUsage  = `GRANT USAGE ON SCHEMA $schema TO cloudsqliamuser;`
	revokeUsage = `REVOKE USAGE ON SCHEMA $schema FROM cloudsqliamuser;`
)

func PrepareAccess(ctx context.Context, appName string, fl *flag.Prepare, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	if _, err := GetSecretValues(ctx, appName, fl.Postgres, ReasonPrepareAccess, out); err != nil {
		return err
	}

	prependUsageIfNotPublic := func(statement string) string {
		if fl.Schema != "public" {
			return grantUsage + "\n" + statement
		}
		return statement
	}

	if fl.AllPrivileges {
		return sqlExecAsAppUser(ctx, appName, fl.Namespace, fl.Context, fl.Schema, prependUsageIfNotPublic(grantAllPrivs))
	} else {
		return sqlExecAsAppUser(ctx, appName, fl.Namespace, fl.Context, fl.Schema, prependUsageIfNotPublic(grantSelectPrivs))
	}
}

func RevokeAccess(ctx context.Context, appName string, fl *flag.Revoke, out *naistrix.OutputWriter) error {
	// Get secret values (access is logged for audit purposes)
	if _, err := GetSecretValues(ctx, appName, fl.Postgres, ReasonRevokeAccess, out); err != nil {
		return err
	}

	q := revokeAllPrivs
	if fl.Schema != "public" {
		q += "\n" + revokeUsage
	}
	return sqlExecAsAppUser(ctx, appName, fl.Namespace, fl.Context, fl.Schema, q)
}

func sqlExecAsAppUser(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, schema, statement string) error {
	dbInfo, err := NewDBInfo(ctx, appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	schema = pq.QuoteIdentifier(schema)
	statement = strings.ReplaceAll(statement, "$schema", schema)
	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ProxyConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, statement)
	if err != nil {
		return formatInvalidGrantError(err)
	}

	return nil
}
