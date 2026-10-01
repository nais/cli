package command

import (
	"slices"
	"testing"
)

func TestPsqlEnvironmentIgnoresInheritedPostgresSettings(t *testing.T) {
	got := withoutPostgresEnv([]string{"PATH=/bin", "PGSERVICE=unsafe", "PGSSLMODE=disable", "PGPASSWORD=old", "HOME=/home/user"})
	want := []string{"PATH=/bin", "HOME=/home/user"}
	if !slices.Equal(got, want) {
		t.Fatalf("environment = %v, want %v", got, want)
	}
}
