package cli_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nais/cli/pkg/cli"
)

func TestApplication_Run(t *testing.T) {
	defer func() {
		contains := "must have at least one command"
		if r := recover(); r == nil {
			t.Fatalf("expected panic for command with no name, but did not panic")
		} else if msg := r.(string); !strings.Contains(msg, contains) {
			t.Fatalf("expected panic message to contain %q, got: %q", contains, msg)
		}
	}()
	_, _ = (&cli.Application{}).Run(context.Background(), cli.Stdout(), []string{})
}
