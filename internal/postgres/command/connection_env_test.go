package command

import (
	"net"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/nais/cli/internal/postgres"
)

func TestPostgresConnectionEnv(t *testing.T) {
	connection := postgres.Connection{
		ServerName: "postgres.internal", Username: "user@example.com", Password: "sensitive-password",
		CACertificate: "sensitive-ca", RelayEndpoint: "sensitive-endpoint", RelayAccess: "sensitive-access", RelayToken: "sensitive-token",
	}
	for _, host := range []string{"127.0.0.1", "::1"} {
		for _, password := range []bool{false, true} {
			env := postgresConnectionEnv(connection, &net.TCPAddr{IP: net.ParseIP(host), Port: 54321}, "/path/to/ca.crt", password)
			want := []string{
				"PGHOST=postgres.internal", "PGHOSTADDR=" + host, "PGPORT=54321", "PGUSER=user@example.com",
				"PGSSLMODE=verify-full", "PGSSLROOTCERT=/path/to/ca.crt", "PGCONNECT_TIMEOUT=10",
			}
			if password {
				want = append(want, "PGPASSWORD=sensitive-password")
			}
			if !slices.Equal(env, want) {
				t.Errorf("environment = %v, want %v", env, want)
			}
		}
	}
}

func TestPostgresEnvOutputShellRoundTrip(t *testing.T) {
	values := []string{"", "plain", "space and\ttab", "apostrophe's", "quotes\"and\\slashes", "$(exit 42); `exit 43` $HOME # comment", "line one\nline two", "equals=value", "æøå"}
	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			output := postgresEnvOutput([]string{"PGPASSWORD=" + value})
			if !strings.HasPrefix(output, "PGPASSWORD='") || !strings.HasSuffix(output, "'\n") {
				t.Fatalf("unexpected assignment: %q", output)
			}
			cmd := exec.Command("sh", "-c", output+`printf '%s' "$PGPASSWORD"`)
			got, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != value {
				t.Errorf("shell value = %q, want %q", got, value)
			}
		})
	}
}
