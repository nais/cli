package validation

import (
	"fmt"
)

var CheckTeam = func(team string) error {
	if team == "" {
		return fmt.Errorf("team cannot be empty, set team using 'nais config set team <team>' or the --team flag")
	}
	return nil
}
