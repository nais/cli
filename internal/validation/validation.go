package validation

import (
	"context"
	"fmt"

	"github.com/nais/naistrix"
)

type teamChecker interface {
	HasTeam() bool
}

// RequireTeam returns a [naistrix.ValidateFunc] that ensures a team is set on the given argument. The argument to this
// function is typically a pointer to a flag struct.
func RequireTeam(f any) naistrix.ValidateFunc {
	return func(context.Context, *naistrix.Arguments) error {
		c, ok := f.(teamChecker)
		if ok && c.HasTeam() {
			return nil
		}

		return fmt.Errorf("missing required team, specify a team using `nais defaults set team <team>` or by using the -t, --team flag")
	}
}

type envChecker interface {
	HasEnvironment() bool
}

// RequireEnvironment returns a [naistrix.ValidateFunc] that ensures an environment is set on the given argument. The
// argument to this function is typically a pointer to a flag struct.
func RequireEnvironment(f any) naistrix.ValidateFunc {
	return func(context.Context, *naistrix.Arguments) error {
		c, ok := f.(envChecker)
		if ok && c.HasEnvironment() {
			return nil
		}

		return fmt.Errorf("missing required environment, specify an environment using `nais defaults set environment <environment>` or by using the -e, --environment flag")
	}
}

// RequireTeamAndEnvironment returns a [naistrix.ValidateFunc] that ensures that the provided argument has both team and
// environment fields set.
func RequireTeamAndEnvironment(f any) naistrix.ValidateFunc {
	return naistrix.ValidateFuncs(
		RequireTeam(f),
		RequireEnvironment(f),
	)
}
