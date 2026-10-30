package command

import (
	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func CloudSQL(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.CloudSQL{
		GlobalFlags: parentFlags,
	}

	return &naistrix.Command{
		PreRunHooks: []naistrix.PreRunHook{auth.GcloudLoginHook(parentFlags)},
		Name:        "cloudsql",
		Title:       "Manage Google Cloud SQL instances.",
		Description: "Manage Google Cloud SQL instances, including listing, user management, password rotation, and direct database access.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			listCommand(flags),
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
	}
}
