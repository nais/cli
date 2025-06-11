package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
	"github.com/nais/cli/internal/postgres/command/flag"
)

var grantAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT ALL ON SEQUENCES TO cloudsqliamuser;
	GRANT ALL ON ALL TABLES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT ALL ON ALL SEQUENCES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT CREATE ON SCHEMA $schema TO cloudsqliamuser;`

var grantSelectPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema GRANT SELECT ON SEQUENCES TO cloudsqliamuser;
	GRANT SELECT ON ALL TABLES IN SCHEMA $schema TO cloudsqliamuser;
	GRANT SELECT ON ALL SEQUENCES IN SCHEMA $schema TO cloudsqliamuser;`

// this is used for all privileges and select, as it covers both cases
var revokeAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON TABLES FROM cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA $schema REVOKE ALL ON SEQUENCES FROM cloudsqliamuser;
	REVOKE ALL ON ALL TABLES IN SCHEMA $schema FROM cloudsqliamuser;
	REVOKE ALL ON ALL SEQUENCES IN SCHEMA $schema FROM cloudsqliamuser;
	REVOKE CREATE ON SCHEMA $schema FROM cloudsqliamuser;`

func PrepareAccess(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, schema string, allPrivs bool) error {
	if allPrivs {
		return sqlExecAsAppUser(ctx, appName, namespace, cluster, schema, grantAllPrivs)
	} else {
		return sqlExecAsAppUser(ctx, appName, namespace, cluster, schema, grantSelectPrivs)
	}
}

func RevokeAccess(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, schema string) error {
	return sqlExecAsAppUser(ctx, appName, namespace, cluster, schema, revokeAllPrivs)
}

func sqlExecAsAppUser(ctx context.Context, appName string, namespace flag.Namespace, cluster flag.Context, schema, statement string) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
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
