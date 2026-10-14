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

// RequireTeam returns a [naistrix.ValidateFunc] that ensures a team is set on the given argument. The argument to this
// function is typically a pointer to a flag struct.
func RequireTeam(f any) naistrix.ValidateFunc {
	return func(context.Context, *naistrix.Arguments) error {
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

// RequireEnvironment returns a [naistrix.ValidateFunc] that ensures an environment is set on the given argument. The
// argument to this function is typically a pointer to a flag struct.
func RequireEnvironment(f any) naistrix.ValidateFunc {
	return func(context.Context, *naistrix.Arguments) error {
		c, ok := f.(envChecker)
		if ok && c.HasEnvironment() {
			return nil
		}

		return ErrMissingEnvironment
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
