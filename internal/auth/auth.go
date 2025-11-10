package auth

import (
	"fmt"

	"github.com/nais/cli/internal/auth/flag"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

func Auth(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Auth{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:        "auth",
		Title:       "Authentication",
		Description: "Commands related to authentication in the nais platform",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			Login(flags),
			Logout(flags),
		},
	}
}

func Deprecated(cmd *naistrix.Command) {
	cmd.Title = fmt.Sprintf("%s (Deprecated, use `nais auth %s`)", cmd.Title, cmd.Name)
}
