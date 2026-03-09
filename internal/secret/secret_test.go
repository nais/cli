package secret

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestFormatDetails(t *testing.T) {
	t.Parallel()

	recentTime := time.Now().Add(-2 * time.Hour)
	recentExpected := LastModified(recentTime).String()

	tests := []struct {
		name     string
		metadata Metadata
		secret   *gql.GetSecretTeamEnvironmentSecret
		want     [][]string
	}{
		{
			name: "all fields present",
			metadata: Metadata{
				TeamSlug:        "my-team",
				EnvironmentName: "dev",
				Name:            "my-secret",
			},
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Name: "my-secret",
				TeamEnvironment: gql.GetSecretTeamEnvironmentSecretTeamEnvironment{
					Environment: gql.GetSecretTeamEnvironmentSecretTeamEnvironmentEnvironment{Name: "dev"},
				},
				LastModifiedAt: recentTime,
				LastModifiedBy: gql.GetSecretTeamEnvironmentSecretLastModifiedByUser{Email: "user@example.com"},
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "my-team"},
				{"Environment", "dev"},
				{"Name", "my-secret"},
				{"Last Modified", recentExpected},
				{"Modified By", "user@example.com"},
			},
		},
		{
			name: "no modification info",
			metadata: Metadata{
				TeamSlug:        "my-team",
				EnvironmentName: "prod",
				Name:            "db-creds",
			},
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Name: "db-creds",
				TeamEnvironment: gql.GetSecretTeamEnvironmentSecretTeamEnvironment{
					Environment: gql.GetSecretTeamEnvironmentSecretTeamEnvironmentEnvironment{Name: "prod"},
				},
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "my-team"},
				{"Environment", "prod"},
				{"Name", "db-creds"},
			},
		},
		{
			name: "modified at set but no modified by",
			metadata: Metadata{
				TeamSlug:        "team-a",
				EnvironmentName: "staging",
				Name:            "api-keys",
			},
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Name: "api-keys",
				TeamEnvironment: gql.GetSecretTeamEnvironmentSecretTeamEnvironment{
					Environment: gql.GetSecretTeamEnvironmentSecretTeamEnvironmentEnvironment{Name: "staging"},
				},
				LastModifiedAt: recentTime,
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "team-a"},
				{"Environment", "staging"},
				{"Name", "api-keys"},
				{"Last Modified", recentExpected},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDetails(tt.metadata, tt.secret)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatDetails() =\n  %v\nwant\n  %v", got, tt.want)
			}
		})
	}
}

func TestFormatKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret *gql.GetSecretTeamEnvironmentSecret
		want   [][]string
	}{
		{
			name: "multiple keys",
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Keys: []string{"DATABASE_URL", "API_KEY", "SECRET_TOKEN"},
			},
			want: [][]string{
				{"Key"},
				{"DATABASE_URL"},
				{"API_KEY"},
				{"SECRET_TOKEN"},
			},
		},
		{
			name:   "no keys",
			secret: &gql.GetSecretTeamEnvironmentSecret{},
			want: [][]string{
				{"Key"},
			},
		},
		{
			name: "single key",
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Keys: []string{"ONLY_KEY"},
			},
			want: [][]string{
				{"Key"},
				{"ONLY_KEY"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatKeys(tt.secret)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatKeys() =\n  %v\nwant\n  %v", got, tt.want)
			}
		})
	}
}

func TestFormatWorkloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret *gql.GetSecretTeamEnvironmentSecret
		want   [][]string
	}{
		{
			name: "with workloads",
			secret: &gql.GetSecretTeamEnvironmentSecret{
				Workloads: gql.GetSecretTeamEnvironmentSecretWorkloadsWorkloadConnection{
					Nodes: []gql.GetSecretTeamEnvironmentSecretWorkloadsWorkloadConnectionNodesWorkload{
						&gql.GetSecretTeamEnvironmentSecretWorkloadsWorkloadConnectionNodesApplication{
							Name:     "my-app",
							Typename: "Application",
						},
						&gql.GetSecretTeamEnvironmentSecretWorkloadsWorkloadConnectionNodesJob{
							Name:     "my-job",
							Typename: "Job",
						},
					},
				},
			},
			want: [][]string{
				{"Name", "Type"},
				{"my-app", "Application"},
				{"my-job", "Job"},
			},
		},
		{
			name:   "no workloads",
			secret: &gql.GetSecretTeamEnvironmentSecret{},
			want: [][]string{
				{"Name", "Type"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatWorkloads(tt.secret)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatWorkloads() =\n  %v\nwant\n  %v", got, tt.want)
			}
		})
	}
}

func TestLastModified_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{name: "zero time", time: time.Time{}, want: ""},
		{name: "seconds ago", time: time.Now().Add(-30 * time.Second), want: "30s"},
		{name: "minutes ago", time: time.Now().Add(-5 * time.Minute), want: "5m"},
		{name: "hours ago", time: time.Now().Add(-3 * time.Hour), want: "3h"},
		{name: "days ago", time: time.Now().Add(-7 * 24 * time.Hour), want: "7d"},
		{name: "years ago", time: time.Now().Add(-400 * 24 * time.Hour), want: "1y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := LastModified(tt.time).String()
			if got != tt.want {
				t.Errorf("LastModified.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLastModified_MarshalJSON(t *testing.T) {
	t.Parallel()

	ts := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC)
	lm := LastModified(ts)

	data, err := json.Marshal(lm)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	want := `"2025-06-15T12:30:00Z"`
	if string(data) != want {
		t.Errorf("MarshalJSON() = %s, want %s", data, want)
	}

	// Zero time should marshal to empty string
	data, err = json.Marshal(LastModified(time.Time{}))
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(data) != `""` {
		t.Errorf("MarshalJSON() zero = %s, want %q", data, "")
	}
}
