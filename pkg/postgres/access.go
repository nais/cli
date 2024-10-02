package postgres

import (
	"context"
	"database/sql"
)

var grantAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO cloudsqliamuser;
	GRANT ALL ON ALL TABLES IN SCHEMA public TO cloudsqliamuser;
	GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO cloudsqliamuser;
	GRANT CREATE ON SCHEMA public TO cloudsqliamuser;`

var grantSelectPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO cloudsqliamuser;
	GRANT SELECT ON ALL TABLES IN SCHEMA public TO cloudsqliamuser;
	GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO cloudsqliamuser;`

// this is used for all privileges and select, as it covers both cases
var revokeAllPrivs = `ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON TABLES FROM cloudsqliamuser;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE ALL ON SEQUENCES FROM cloudsqliamuser;
	REVOKE ALL ON ALL TABLES IN SCHEMA public FROM cloudsqliamuser;
	REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM cloudsqliamuser;
	REVOKE CREATE ON SCHEMA public FROM cloudsqliamuser;`

func PrepareAccess(ctx context.Context, appName, namespace, cluster string, allPrivs bool) error {
	if allPrivs {
		return sqlExecAsAppUser(ctx, appName, namespace, cluster, grantAllPrivs)
	} else {
		return sqlExecAsAppUser(ctx, appName, namespace, cluster, grantSelectPrivs)
	}
}

func RevokeAccess(ctx context.Context, appName, namespace, cluster string) error {
	return sqlExecAsAppUser(ctx, appName, namespace, cluster, revokeAllPrivs)
}

func sqlExecAsAppUser(ctx context.Context, appName, namespace, cluster, statement string) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

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
