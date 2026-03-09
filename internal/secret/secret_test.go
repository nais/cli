package secret

import (
	"reflect"
	"testing"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestFormatDetails(t *testing.T) {
	t.Parallel()

	modifiedAt := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC)

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
				LastModifiedAt: modifiedAt,
				LastModifiedBy: gql.GetSecretTeamEnvironmentSecretLastModifiedByUser{Email: "user@example.com"},
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "my-team"},
				{"Environment", "dev"},
				{"Name", "my-secret"},
				{"Last Modified", "2025-06-15T12:30:00Z"},
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
				LastModifiedAt: modifiedAt,
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "team-a"},
				{"Environment", "staging"},
				{"Name", "api-keys"},
				{"Last Modified", "2025-06-15T12:30:00Z"},
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
