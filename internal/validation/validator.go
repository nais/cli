package validation

import (
	"context"
	"fmt"

	"github.com/nais/naistrix"
)

var CheckTeam = func(team string) error {
	if team == "" {
		return fmt.Errorf("team cannot be empty, set team using 'nais config set team <team>' or the --team flag")
	}
	return nil
}

func TeamValidator(team string) naistrix.ValidateFunc {
	return func(_ context.Context, _ *naistrix.Arguments) error {
		return CheckTeam(team)
	}
}
