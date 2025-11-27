package validation

import (
	"fmt"
)

var CheckEnvironment = func(env string) error {
	if env == "" {
		return fmt.Errorf("environment cannot be empty, set environment using --environment/-e flag")
	}
	return nil
}

var CheckTeam = func(team string) error {
	if team == "" {
		return fmt.Errorf("team cannot be empty, set team using 'nais config set team <team>' or the --team flag")
	}
	return nil
}
