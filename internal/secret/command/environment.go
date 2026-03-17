package command

import (
	"fmt"
	"os"

	"github.com/nais/cli/internal/cliflags"
)

func validateSingleEnvironmentFlagUsage() error {
	if countEnvironmentFlagsInCLIArgs() > 1 {
		return fmt.Errorf("only one --environment/-e flag may be provided")
	}
	return nil
}

func countEnvironmentFlagsInCLIArgs() int {
	return cliflags.CountFlagOccurrences(os.Args, "-e", "--environment")
}

func environmentValuesFromCLIArgs() []string {
	return cliflags.UniqueFlagValues(os.Args, "-e", "--environment")
}
