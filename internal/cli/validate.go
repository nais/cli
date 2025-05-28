package cli

import (
	"context"
	"fmt"
)

// ValidateExactArgs checks that the user has provided an exact amount of arguments to the command.
func ValidateExactArgs(n int) ValidateFunc {
	return func(_ context.Context, args []string) error {
		if len(args) != n {
			return fmt.Errorf("expected exactly %d arguments, got %d", n, len(args))
		}

		return nil
	}
}

// ValidateMinArgs checks that the user has provided a minimum amount of arguments to the command.
func ValidateMinArgs(n int) ValidateFunc {
	return func(_ context.Context, args []string) error {
		if len(args) < n {
			return fmt.Errorf("expected at least %d arguments, got %d", n, len(args))
		}

		return nil
	}
}
