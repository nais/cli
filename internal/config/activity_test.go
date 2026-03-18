package config

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildConfigActivity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name         string
		resources    []configActivityResource
		configName   string
		environments []string
		wantFound    bool
		want         []ConfigActivity
	}{
		{
			name:       "exact name match with fallback environment",
			configName: "app-settings",
			resources: []configActivityResource{
				{
					Name:           "app-settings",
					DefaultEnvName: "dev-gcp",
					Entries: []configActivityEntry{
						{CreatedAt: now, Actor: "alice@example.com", Message: "Updated config value", EnvironmentName: ""},
					},
				},
			},
			wantFound: true,
			want: []ConfigActivity{
				{CreatedAt: now, Actor: "alice@example.com", Environment: "dev-gcp", Message: "Updated config value"},
			},
		},
		{
			name:       "entry with explicit environment overrides default",
			configName: "app-settings",
			resources: []configActivityResource{
				{
					Name:           "app-settings",
					DefaultEnvName: "dev-gcp",
					Entries: []configActivityEntry{
						{CreatedAt: now, Actor: "bob@example.com", Message: "Created config", EnvironmentName: "prod-gcp"},
					},
				},
			},
			wantFound: true,
			want: []ConfigActivity{
				{CreatedAt: now, Actor: "bob@example.com", Environment: "prod-gcp", Message: "Created config"},
			},
		},
		{
			name:       "not found in requested environment",
			configName: "app-settings",
			resources: []configActivityResource{
				{
					Name:           "app-settings",
					DefaultEnvName: "dev-gcp",
					Entries: []configActivityEntry{
						{CreatedAt: now, Actor: "alice@example.com", Message: "Updated config value", EnvironmentName: ""},
					},
				},
			},
			environments: []string{"prod-gcp"},
			wantFound:    false,
			want:         []ConfigActivity{},
		},
		{
			name:       "not found when only partial name exists",
			configName: "app",
			resources: []configActivityResource{
				{
					Name:           "app-settings",
					DefaultEnvName: "dev-gcp",
				},
			},
			wantFound: false,
			want:      []ConfigActivity{},
		},
		{
			name:       "multiple resources same name different envs",
			configName: "db-config",
			resources: []configActivityResource{
				{
					Name:           "db-config",
					DefaultEnvName: "dev-gcp",
					Entries: []configActivityEntry{
						{CreatedAt: now, Actor: "alice@example.com", Message: "Added key", EnvironmentName: ""},
					},
				},
				{
					Name:           "db-config",
					DefaultEnvName: "prod-gcp",
					Entries: []configActivityEntry{
						{CreatedAt: now, Actor: "bob@example.com", Message: "Updated key", EnvironmentName: ""},
					},
				},
			},
			environments: []string{"prod-gcp"},
			wantFound:    true,
			want: []ConfigActivity{
				{CreatedAt: now, Actor: "bob@example.com", Environment: "prod-gcp", Message: "Updated key"},
			},
		},
		{
			name:       "empty resources",
			configName: "anything",
			resources:  nil,
			wantFound:  false,
			want:       []ConfigActivity{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := buildConfigActivity(tt.resources, tt.configName, tt.environments)
			if found != tt.wantFound {
				t.Fatalf("buildConfigActivity() found = %v, want %v", found, tt.wantFound)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildConfigActivity() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
