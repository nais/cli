package command

import (
	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Auth(parentFlags *flags.GlobalFlags) *naistrix.Command {
	f := &flag.Auth{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "auth",
		Title:       "Authentication",
		Description: "Commands related to authentication in the nais platform",
		StickyFlags: f,
		SubCommands: []*naistrix.Command{
			login(f),
			logout(f),
			printAccessToken(f),
			workloadIdentityMetadata(f),
		},
	}
}
