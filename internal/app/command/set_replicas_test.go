package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/validation"
)

func TestSetReplicasFlagValidation(t *testing.T) {
	for _, tt := range []struct {
		name               string
		environment        flags.Environment
		min, max           int
		wantError          string
		missingEnvironment bool
	}{
		{name: "missing environment", missingEnvironment: true},
		{name: "environment checked before replica counts", min: -1, max: -1, missingEnvironment: true},
		{name: "negative minimum", environment: "dev", min: -1, max: 1, wantError: "--min must be a non-negative integer"},
		{name: "negative maximum", environment: "dev", max: -1, wantError: "--max must be a non-negative integer"},
		{name: "minimum exceeds maximum", environment: "dev", min: 2, max: 1, wantError: "--min (2) cannot be greater than --max (1)"},
		{name: "valid counts", environment: "dev", min: 1, max: 2},
		{name: "zero counts", environment: "dev"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parent := &flag.App{GlobalFlags: &flags.GlobalFlags{
				AdditionalFlags: &flags.AdditionalFlags{Team: "my-team", Environment: tt.environment},
			}}
			cmd := setReplicas(parent)
			f := cmd.Flags.(*flag.SetReplicas)
			f.Min, f.Max = tt.min, tt.max
			err := cmd.ValidateFlagsFunc(t.Context())
			if tt.missingEnvironment {
				if !errors.Is(err, validation.ErrMissingEnvironment) {
					t.Errorf("error = %v, want ErrMissingEnvironment", err)
				}
			} else if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
