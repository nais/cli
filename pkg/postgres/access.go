package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

var prepareDdlStatements = []string{
	"alter default privileges in schema public grant %s on tables to cloudsqliamuser;",
	"alter default privileges in schema public grant %s on sequences to cloudsqliamuser;",
	"grant %s on all tables in schema public to cloudsqliamuser;",
	"grant %s on all sequences in schema public to cloudsqliamuser;",
}

var grantCreatePublicStatements = []string{
	"alter default privileges in schema public grant create to cloudsqliamuser;",
	"grant create in schema public to cloudsqliamuser;",
}

func PrepareAccess(ctx context.Context, appName, namespace, cluster, database string, allPrivs bool) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster, database)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	for _, ddl := range prepareDdlStatements {
		grant := "SELECT"
		if allPrivs {
			grant = "ALL"
		}
		_, err = db.ExecContext(ctx, fmt.Sprint(ddl, grant))
		if err != nil {
			return err
		}
	}
	if allPrivs {
		for _, stmt := range grantCreatePublicStatements {
			_, err = db.ExecContext(ctx, stmt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var revokeDdlStatements = []string{
	"alter default privileges in schema public revoke ALL on tables from cloudsqliamuser;",
	"alter default privileges in schema public revoke ALL on sequences from cloudsqliamuser;",
	"revoke ALL on all tables in schema public from cloudsqliamuser;",
	"revoke ALL on all sequences in schema public from cloudsqliamuser;",
}

func RevokeAccess(ctx context.Context, appName, namespace, cluster, database string) error {
	dbInfo, err := NewDBInfo(appName, namespace, cluster, database)
	if err != nil {
		return err
	}
	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	db, err := sql.Open("cloudsqlpostgres", connectionInfo.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	for _, ddl := range revokeDdlStatements {
		_, err = db.ExecContext(ctx, ddl)
		if err != nil {
			return formatInvalidGrantError(err)
		}
	}

	return nil
}
