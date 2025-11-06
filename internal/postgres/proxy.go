package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func RunProxy(ctx context.Context, appName string, cluster flag.Context, namespace flag.Namespace, host string, port uint, verbose bool, out *naistrix.OutputWriter) error {
	dbInfo, err := NewDBInfo(ctx, appName, namespace, cluster)
	if err != nil {
		return err
	}

	return dbInfo.RunProxy(ctx, host, &port, make(chan<- int, 1), out, true)
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}

func checkPostgresqlPassword(out *naistrix.OutputWriter) error {
	if _, ok := os.LookupEnv("PGPASSWORD"); ok {
		return fmt.Errorf("PGPASSWORD is set, please unset it before running this command")
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		out.Println("could not get home directory, can not check for .pgpass file")
		return nil
	}

	if s, err := os.Stat(filepath.Join(dirname, ".pgpass")); err == nil && !s.IsDir() {
		return fmt.Errorf("found .pgpass file in home directory, please remove it before running this command")
	}
	return nil
}
