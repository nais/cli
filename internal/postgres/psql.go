package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nais/cli/internal/cli"
)

func RunPSQL(ctx context.Context, appName, cluster, namespace string, verbose bool, out cli.Output) error {
	psqlPath, err := exec.LookPath("psql")
	if err != nil {
		return err
	}

	dbInfo, err := NewDBInfo(appName, namespace, cluster)
	if err != nil {
		return err
	}

	projectID, err := dbInfo.ProjectID(ctx)
	if err != nil {
		return err
	}

	connectionName, err := dbInfo.ConnectionName(ctx)
	if err != nil {
		return err
	}

	connectionInfo, err := dbInfo.DBConnection(ctx)
	if err != nil {
		return err
	}

	email, err := currentEmail(ctx)
	if err != nil {
		return err
	}

	portCh := make(chan int, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err := runProxy(ctx, projectID, connectionName, "localhost:0", portCh, verbose, out)
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
		"--username", email,
		"--dbname", connectionInfo.dbName,
	}

	cmd := exec.CommandContext(ctx, psqlPath, arguments...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}
