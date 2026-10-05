package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/gcloud"
	"github.com/nais/naistrix"
)

func CloudSQL(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.CloudSQL{
		GlobalFlags: parentFlags,
	}

	return &naistrix.Command{
		Name: "cloudsql",
		// TODO: Remove the aliases once users have moved from the old `nais postgres` (Cloud SQL) commands.
		Aliases:     []string{"postgres", "pg"},
		Title:       "Manage Google Cloud SQL instances.",
		Description: "Manage Google Cloud SQL instances, including listing, migration, user management, password rotation, and direct database access.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listCommand(flags),
			migrateCommand(flags),
			passwordCommand(flags),
			usersCommand(flags),
			enableAuditCommand(flags),
			verifyAuditCommand(flags),
			grantCommand(flags),
			prepareCommand(flags),
			proxyCommand(flags),
			psqlCommand(flags),
			revokeCommand(flags),
		},
		ValidateFunc: func(ctx context.Context, _ *naistrix.Arguments) error {
			warnIfLegacyAlias(os.Args[1:], os.Stderr)
			_, err := gcloud.ValidateAndGetUserLogin(ctx, false)
			return err
		},
	}
}

// warnIfLegacyAlias tells users who typed `nais postgres` or `nais pg` that these now mean Cloud SQL.
// naistrix does not expose which alias was used, so the first non-flag argument is inspected.
func warnIfLegacyAlias(args []string, w io.Writer) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if arg == "postgres" || arg == "pg" {
			_, _ = fmt.Fprintf(w, "Warning: nais %s has moved to nais cloudsql. Use nais cloudsql instead, as nais %[1]s will stop working for Cloud SQL in the future.\n", arg)
		}
		return
	}
}
