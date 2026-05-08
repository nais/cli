package validation

import (
	"context"
	"testing"

	"github.com/nais/cli/internal/flags"
)

func TestRequireTeam(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "valid AdditionalFlags with team",
			input:   &flags.AdditionalFlags{Team: "my-team"},
			wantErr: false,
		},
		{
			name:    "AdditionalFlags with empty team",
			input:   &flags.AdditionalFlags{Team: ""},
			wantErr: true,
		},
		{
			name:    "AdditionalFlags zero value",
			input:   &flags.AdditionalFlags{},
			wantErr: true,
		},
		{
			name:    "non-flag type (string)",
			input:   "my-team",
			wantErr: true,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			name: "GlobalFlags embedding *AdditionalFlags with team",
			input: &flags.GlobalFlags{
				AdditionalFlags: &flags.AdditionalFlags{Team: "my-team"},
			},
			wantErr: false,
		},
		{
			name: "GlobalFlags embedding *AdditionalFlags without team",
			input: &flags.GlobalFlags{
				AdditionalFlags: &flags.AdditionalFlags{},
			},
			wantErr: true,
		},
		{
			name: "GlobalFlags with nil embedded *AdditionalFlags",
			input: &flags.GlobalFlags{
				AdditionalFlags: nil,
			},
			wantErr: true,
		},
		{
			name: "struct embedding AdditionalFlags by value with team",
			input: struct {
				*flags.AdditionalFlags
			}{
				AdditionalFlags: &flags.AdditionalFlags{Team: "my-team"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := RequireTeam(tt.input)
			err := validate(context.Background(), nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireTeam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireTeamAndEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "AdditionalFlags with team and environment",
			input:   &flags.AdditionalFlags{Team: "my-team", Environment: "dev"},
			wantErr: false,
		},
		{
			name:    "AdditionalFlags with team but no environment",
			input:   &flags.AdditionalFlags{Team: "my-team"},
			wantErr: true,
		},
		{
			name:    "AdditionalFlags with environment but no team",
			input:   &flags.AdditionalFlags{Environment: "dev"},
			wantErr: true,
		},
		{
			name:    "AdditionalFlags zero value",
			input:   &flags.AdditionalFlags{},
			wantErr: true,
		},
		{
			name:    "non-flag type (string)",
			input:   "my-team",
			wantErr: true,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := RequireTeamAndEnvironment(tt.input)
			err := validate(context.Background(), nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireTeamAndEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
