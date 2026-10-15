package validation

import (
	"context"
	"errors"

	"github.com/nais/naistrix"
)

var (
	ErrMissingTeam        = errors.New("missing required team, specify -t, --team or set a default team")
	ErrMissingEnvironment = errors.New("missing required environment, specify -e, --environment or set a default environment")
)

type teamChecker interface {
	HasTeam() bool
}

// RequireTeamFlag returns a [naistrix.ValidateFlagsFunc] that ensures a team is set on the given argument. The argument
// to this function is typically a pointer to a flag struct.
func RequireTeamFlag(f any) naistrix.ValidateFlagsFunc {
	return func(context.Context) error {
		c, ok := f.(teamChecker)
		if ok && c.HasTeam() {
			return nil
		}

		return ErrMissingTeam
	}
}

type envChecker interface {
	HasEnvironment() bool
}

// RequireEnvironmentFlag returns a [naistrix.ValidateFlagsFunc] that ensures an environment is set on the given
// argument. The argument to this function is typically a pointer to a flag struct.
func RequireEnvironmentFlag(f any) naistrix.ValidateFlagsFunc {
	return func(context.Context) error {
		c, ok := f.(envChecker)
		if ok && c.HasEnvironment() {
			return nil
		}

		return ErrMissingEnvironment
	}
}

// RequireTeamAndEnvironmentFlags returns a [naistrix.ValidateFlagsFunc] that ensures that the provided argument has
// both team and environment fields set.
func RequireTeamAndEnvironmentFlags(f any) naistrix.ValidateFlagsFunc {
	return naistrix.ValidateFlagsFuncs(
		RequireTeamFlag(f),
		RequireEnvironmentFlag(f),
	)
}
