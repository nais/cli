package config

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
		config   *gql.GetConfigTeamEnvironmentConfig
		want     [][]string
	}{
		{
			name: "all fields present",
			metadata: Metadata{
				TeamSlug:        "my-team",
				EnvironmentName: "dev",
				Name:            "my-config",
			},
			config: &gql.GetConfigTeamEnvironmentConfig{
				Name: "my-config",
				TeamEnvironment: gql.GetConfigTeamEnvironmentConfigTeamEnvironment{
					Environment: gql.GetConfigTeamEnvironmentConfigTeamEnvironmentEnvironment{Name: "dev"},
				},
				LastModifiedAt: recentTime,
				LastModifiedBy: gql.GetConfigTeamEnvironmentConfigLastModifiedByUser{Email: "user@example.com"},
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "my-team"},
				{"Environment", "dev"},
				{"Name", "my-config"},
				{"Last Modified", recentExpected},
				{"Modified By", "user@example.com"},
			},
		},
		{
			name: "no modification info",
			metadata: Metadata{
				TeamSlug:        "my-team",
				EnvironmentName: "prod",
				Name:            "db-config",
			},
			config: &gql.GetConfigTeamEnvironmentConfig{
				Name: "db-config",
				TeamEnvironment: gql.GetConfigTeamEnvironmentConfigTeamEnvironment{
					Environment: gql.GetConfigTeamEnvironmentConfigTeamEnvironmentEnvironment{Name: "prod"},
				},
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "my-team"},
				{"Environment", "prod"},
				{"Name", "db-config"},
			},
		},
		{
			name: "modified at set but no modified by",
			metadata: Metadata{
				TeamSlug:        "team-a",
				EnvironmentName: "staging",
				Name:            "app-settings",
			},
			config: &gql.GetConfigTeamEnvironmentConfig{
				Name: "app-settings",
				TeamEnvironment: gql.GetConfigTeamEnvironmentConfigTeamEnvironment{
					Environment: gql.GetConfigTeamEnvironmentConfigTeamEnvironmentEnvironment{Name: "staging"},
				},
				LastModifiedAt: recentTime,
			},
			want: [][]string{
				{"Field", "Value"},
				{"Team", "team-a"},
				{"Environment", "staging"},
				{"Name", "app-settings"},
				{"Last Modified", recentExpected},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDetails(tt.metadata, tt.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatDetails() =\n  %v\nwant\n  %v", got, tt.want)
			}
		})
	}
}

func TestFormatData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []gql.GetConfigTeamEnvironmentConfigValuesConfigValue
		want   [][]string
	}{
		{
			name: "multiple values",
			values: []gql.GetConfigTeamEnvironmentConfigValuesConfigValue{
				{Name: "DATABASE_HOST", Value: "db.example.com"},
				{Name: "LOG_LEVEL", Value: "info"},
				{Name: "PORT", Value: "8080"},
			},
			want: [][]string{
				{"Key", "Value"},
				{"DATABASE_HOST", "db.example.com"},
				{"LOG_LEVEL", "info"},
				{"PORT", "8080"},
			},
		},
		{
			name:   "no values",
			values: nil,
			want: [][]string{
				{"Key", "Value"},
			},
		},
		{
			name: "single value",
			values: []gql.GetConfigTeamEnvironmentConfigValuesConfigValue{
				{Name: "ONLY_KEY", Value: "only_value"},
			},
			want: [][]string{
				{"Key", "Value"},
				{"ONLY_KEY", "only_value"},
			},
		},
		{
			name: "value with empty string",
			values: []gql.GetConfigTeamEnvironmentConfigValuesConfigValue{
				{Name: "EMPTY_KEY", Value: ""},
			},
			want: [][]string{
				{"Key", "Value"},
				{"EMPTY_KEY", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatData(tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatData() =\n  %v\nwant\n  %v", got, tt.want)
			}
		})
	}
}

func TestFormatWorkloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *gql.GetConfigTeamEnvironmentConfig
		want   [][]string
	}{
		{
			name: "with workloads",
			config: &gql.GetConfigTeamEnvironmentConfig{
				Workloads: gql.GetConfigTeamEnvironmentConfigWorkloadsWorkloadConnection{
					Nodes: []gql.GetConfigTeamEnvironmentConfigWorkloadsWorkloadConnectionNodesWorkload{
						&gql.GetConfigTeamEnvironmentConfigWorkloadsWorkloadConnectionNodesApplication{
							Name:     "my-app",
							Typename: "Application",
						},
						&gql.GetConfigTeamEnvironmentConfigWorkloadsWorkloadConnectionNodesJob{
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
			config: &gql.GetConfigTeamEnvironmentConfig{},
			want: [][]string{
				{"Name", "Type"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatWorkloads(tt.config)
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
