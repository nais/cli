package cli

import (
	"context"
	"fmt"
)

func ValidateExactArgs(n int) ValidateFunc {
	return func(_ context.Context, args []string) error {
		if len(args) != n {
			return fmt.Errorf("expected exactly %d arguments, got %d", n, len(args))
		}

		return nil
	}
}

func ValidateMinArgs(n int) ValidateFunc {
	return func(_ context.Context, args []string) error {
		if len(args) < n {
			return fmt.Errorf("expected at least %d arguments, got %d", n, len(args))
		}

		return nil
	}
}
