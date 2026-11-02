package command

import (
	"fmt"
	"net"
	"strings"

	"github.com/nais/cli/internal/postgres"
)

// postgresConnectionEnv is shared by psql and the machine-readable proxy output so both use the same end-to-end TLS
// settings. Relay credentials are never exposed.
func postgresConnectionEnv(connection postgres.Connection, address *net.TCPAddr, caPath string, includePassword bool) []string {
	env := []string{
		"PGHOST=" + connection.ServerName,
		"PGHOSTADDR=" + address.IP.String(),
		fmt.Sprintf("PGPORT=%d", address.Port),
		"PGUSER=" + connection.Username,
		"PGSSLMODE=verify-full",
		"PGSSLROOTCERT=" + caPath,
		"PGCONNECT_TIMEOUT=10",
	}
	if includePassword {
		env = append(env, "PGPASSWORD="+connection.Password)
	}
	return env
}

func postgresEnvOutput(env []string) string {
	var output strings.Builder
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		// POSIX shell single quoting prevents expansion and command substitution,
		// including when credentials contain whitespace or shell metacharacters.
		_, _ = fmt.Fprintf(&output, "%s='%s'\n", key, strings.ReplaceAll(value, "'", "'\"'\"'"))
	}
	return output.String()
}
