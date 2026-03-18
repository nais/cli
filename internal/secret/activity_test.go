package secret

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildSecretActivity(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name         string
		resources    []secretActivityResource
		secretName   string
		environments []string
		wantFound    bool
		want         []SecretActivity
	}{
		{
			name:       "exact name match with fallback environment",
			secretName: "db-credentials",
			resources: []secretActivityResource{
				{
					Name:           "db-credentials",
					DefaultEnvName: "dev-gcp",
					Entries: []secretActivityEntry{
						{CreatedAt: now, Actor: "alice@example.com", Message: "Viewed secret values", EnvironmentName: ""},
					},
				},
			},
			wantFound: true,
			want: []SecretActivity{
				{CreatedAt: now, Actor: "alice@example.com", Environment: "dev-gcp", Message: "Viewed secret values"},
			},
		},
		{
			name:       "not found in requested environment",
			secretName: "db-credentials",
			resources: []secretActivityResource{
				{
					Name:           "db-credentials",
					DefaultEnvName: "dev-gcp",
					Entries: []secretActivityEntry{
						{CreatedAt: now, Actor: "alice@example.com", Message: "Viewed secret values", EnvironmentName: ""},
					},
				},
			},
			environments: []string{"prod-gcp"},
			wantFound:    false,
			want:         []SecretActivity{},
		},
		{
			name:       "not found when only partial name exists",
			secretName: "db",
			resources: []secretActivityResource{
				{
					Name:           "db-credentials",
					DefaultEnvName: "dev-gcp",
				},
			},
			wantFound: false,
			want:      []SecretActivity{},
		},
		{
			name:       "sorted by created time descending",
			secretName: "db-credentials",
			resources: []secretActivityResource{
				{
					Name:           "db-credentials",
					DefaultEnvName: "dev-gcp",
					Entries: []secretActivityEntry{
						{CreatedAt: now.Add(-2 * time.Hour), Actor: "alice@example.com", Message: "Older event", EnvironmentName: ""},
					},
				},
				{
					Name:           "db-credentials",
					DefaultEnvName: "prod-gcp",
					Entries: []secretActivityEntry{
						{CreatedAt: now, Actor: "bob@example.com", Message: "Newest event", EnvironmentName: ""},
					},
				},
			},
			wantFound: true,
			want: []SecretActivity{
				{CreatedAt: now, Actor: "bob@example.com", Environment: "prod-gcp", Message: "Newest event"},
				{CreatedAt: now.Add(-2 * time.Hour), Actor: "alice@example.com", Environment: "dev-gcp", Message: "Older event"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := buildSecretActivity(tt.resources, tt.secretName, tt.environments)
			if found != tt.wantFound {
				t.Fatalf("buildSecretActivity() found = %v, want %v", found, tt.wantFound)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildSecretActivity() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
