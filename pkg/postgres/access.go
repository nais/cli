package postgres

import (
	"context"
	"database/sql"
	"strings"
)

var prepareDdlStatements = []string{
	"alter default privileges in schema public grant CHANGEME on tables to cloudsqliamuser;",
	"alter default privileges in schema public grant CHANGEME on sequences to cloudsqliamuser;",
	"grant CHANGEME on all tables in schema public to cloudsqliamuser;",
	"grant CHANGEME on all sequences in schema public to cloudsqliamuser;",
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
		_, err = db.ExecContext(ctx, setGrant(ddl, allPrivs))
		if err != nil {
			return err
		}
	}

	return nil
}

func setGrant(sql string, allPrivs bool) string {
	sqlGrant := "SELECT"
	if allPrivs {
		sqlGrant = "ALL"
	}
	return strings.Replace(sql, "CHANGEME", sqlGrant, 1)
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
