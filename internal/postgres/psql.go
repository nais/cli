package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func RunPSQL(ctx context.Context, appName string, fl *flag.Psql, out *naistrix.OutputWriter) error {
	// Get secret values with user-provided reason (access is logged for audit purposes)
	if _, err := GetSecretValuesWithUserReason(ctx, appName, fl.Postgres, fl.Reason, out); err != nil {
		return err
	}

	psqlPath, err := exec.LookPath("psql")
	if err != nil {
		return err
	}

	dbInfo, err := NewDBInfo(ctx, appName, fl.Namespace, fl.Context)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	portCh := make(chan int, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err := dbInfo.RunProxy(ctx, "localhost", nil, portCh, out, false)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			out.Printf("ERROR: %v", err)
			cancel()
		}
	}()
	port := <-portCh

	out.Printf("Running proxy on localhost:%v\n", port)

	arguments := []string{
		"--host", "localhost",
		"--port", fmt.Sprintf("%d", port),
		"--username", connectionInfo.email,
		"--dbname", connectionInfo.dbName,
	}

	cmd := exec.CommandContext(ctx, psqlPath, arguments...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("PGPASSWORD=%s", connectionInfo.password))
	cmd.Env = environ

	return cmd.Run()
}
