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
			name:       "found true but filtered out by environment",
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
			wantFound:    true,
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
