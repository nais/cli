package migrate_test

import (
	"testing"

	"github.com/nais/cli/internal/option"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/config"
)

func TestCommand(t *testing.T) {
	tests := map[string]struct {
		mutateFn func(cfg *config.Config)
		expected string
	}{
		"happy path with reasonable lengths for app and instance": {
			mutateFn: func(cfg *config.Config) {},
			expected: "migration-some-app-target-instance-setup",
		},
		"very long app name": {
			mutateFn: func(cfg *config.Config) {
				cfg.AppName = "some-unnecessarily-long-app-name-that-should-be-truncated"
			},
			expected: "migration-some-unnecessarily-long-app-eb4938d8-setup",
		},
		"very long instance name": {
			mutateFn: func(cfg *config.Config) {
				cfg.Target.InstanceName = option.Some("some-unnecessarily-long-instance-name-that-should-be-truncated")
			},
			expected: "migration-some-app-some-unnecessarily-63093bcb-setup",
		},
	}

	const cmd = migrate.CommandSetup

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.Config{
				AppName:   "some-app",
				Namespace: "test-namespace",
				Target: config.InstanceConfig{
					InstanceName: option.Some("target-instance"),
				},
			}
			tc.mutateFn(&cfg)

			actual := cmd.JobName(cfg)
			if len(actual) > 52 {
				t.Errorf("job name exceeds 52 characters: %s", actual)
			}
			if actual != tc.expected {
				t.Errorf("expected job name %q, got %q", tc.expected, actual)
			}
		})
	}
}
