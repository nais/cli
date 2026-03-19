package command

import "testing"

func TestSelectConfigEnvironment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		team      string
		config    string
		provided  string
		envs      []string
		wantEnv   string
		wantError string
	}{
		{
			name:     "provided environment exists",
			team:     "nais",
			config:   "my-config",
			provided: "dev-gcp",
			envs:     []string{"dev-gcp", "prod-gcp"},
			wantEnv:  "dev-gcp",
		},
		{
			name:      "provided environment missing with alternatives",
			team:      "nais",
			config:    "my-config",
			provided:  "staging-gcp",
			envs:      []string{"prod-gcp", "dev-gcp"},
			wantError: "config \"my-config\" does not exist in environment \"staging-gcp\"; available environments: dev-gcp, prod-gcp",
		},
		{
			name:      "provided environment missing and config absent",
			team:      "nais",
			config:    "my-config",
			provided:  "dev-gcp",
			envs:      nil,
			wantError: "config \"my-config\" not found in team \"nais\"",
		},
		{
			name:      "no provided and no environments",
			team:      "nais",
			config:    "my-config",
			envs:      nil,
			wantError: "config \"my-config\" not found in team \"nais\"",
		},
		{
			name:    "no provided and one environment",
			team:    "nais",
			config:  "my-config",
			envs:    []string{"dev-gcp"},
			wantEnv: "dev-gcp",
		},
		{
			name:      "no provided and multiple environments",
			team:      "nais",
			config:    "my-config",
			envs:      []string{"prod-gcp", "dev-gcp"},
			wantError: "config \"my-config\" exists in multiple environments (dev-gcp, prod-gcp); specify -e, --environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotEnv, err := selectConfigEnvironment(tt.team, tt.config, tt.provided, tt.envs)
			if tt.wantError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantError)
				}
				if err.Error() != tt.wantError {
					t.Fatalf("error = %q, want %q", err.Error(), tt.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotEnv != tt.wantEnv {
				t.Fatalf("env = %q, want %q", gotEnv, tt.wantEnv)
			}
		})
	}
}
