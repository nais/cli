package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/naistrix"
)

// legacyCommand warns about the new command path without intercepting the old execution.
// naistrix.Deprecated cannot be used here: it replaces RunFunc instead of running it.
func legacyCommand(path, replacement string, cmd *naistrix.Command) *naistrix.Command {
	run := cmd.RunFunc
	cmd.Title += " (DEPRECATED)"
	cmd.RunFunc = func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
		warning := fmt.Sprintf("Warning: nais %s is deprecated and will be removed in a future release", path)
		if replacement != "" {
			warning += "; use " + replacement + " instead"
		}
		if strings.HasSuffix(path, " grant") || strings.HasSuffix(path, " prepare") || strings.HasSuffix(path, " revoke") || strings.HasSuffix(path, " proxy") || strings.HasSuffix(path, " psql") {
			warning += "; the legacy in-cluster access path will be removed"
		}
		_, _ = fmt.Fprintln(os.Stderr, warning+".")
		return run(ctx, args, out)
	}
	return cmd
}
